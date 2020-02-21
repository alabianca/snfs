package util

import (
	"archive/tar"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// MyIP returns the local non loopback ip address
// given the network (ipv4 | ipv6)
func MyIP(network string) (net.IP, error) {

	addrs, _ := net.InterfaceAddrs()

	for _, addr := range addrs {
		var ip net.IP

		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP

		case *net.IPAddr:
			ip = v.IP
		}

		if ip.IsLoopback() {
			continue
		}

		isIPV6 := strings.Contains(ip.String(), ":")
		if isIPV6 && network == "ipv6" {
			return ip, nil
		}

		if !isIPV6 && network == "ipv4" {
			return ip, nil
		}

	}

	return nil, fmt.Errorf("No IP found")

}

func WriteTarball(writer io.Writer, dir string) (int64, error) {
	tw := tar.NewWriter(writer)

	defer tw.Close()

	// walk path
	var bytesWritten int64
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = path
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		defer f.Close()
		if err != nil {
			return err
		}

		var n int64
		if n, err = io.Copy(tw, f); err != nil {
			return err
		}
		bytesWritten+=n

		return nil
	})

	return bytesWritten, err
}

// ReadTarball reads from reader and creates the resulting directory at target
func ReadTarball(reader io.Reader, target string) error {

	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue

		}

		target := filepath.Join(target, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			//it is a directory. create it if it does not exist
			log.Println("Creating directory...")
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			//regular file. create it
			f, err := os.Create(header.Name)
			//f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}

func RandomID(p []byte) {
	rand.Read(p)
}

func BytesFromHex(in string) ([]byte, error) {
	src := []byte(in)
	out := make([]byte, hex.DecodedLen(len(src)))

	n, err := hex.Decode(out, src)
	if err != nil {
		return nil, err
	}

	return out[:n], nil

}

func SetEnv(key string, val string) error {
	return os.Setenv(key, val)
}
