package util

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
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

func WriteTarball(writer io.Writer, dir string) error {
	tw := tar.NewWriter(writer)

	defer tw.Close()

	// walk path
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}

func ReadTarball(reader io.Reader) error {
	tr := tar.NewReader(reader)
	log.Println("Reading tarball")
	for {
		header, err := tr.Next()
		log.Println(err)
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		io.Copy(os.Stdout, tr)
	}
}
