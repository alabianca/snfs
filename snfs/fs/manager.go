package fs

import (
	"archive/tar"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const dir = "/Users/alexander/go/src/github.com/alabianca/snfs"

// Manager maintains a list of files that are currently
// shared in the local network
type Manager struct {
	files []string
}

// NewManager returns a Manager with zero files
func NewManager() *Manager {
	return &Manager{
		files: make([]string, 0),
	}
}

// Add adds a file to the list of files
func (m *Manager) Add(fname string) {
	m.files = append(m.files, fname)
}

// Shutdown removes all the files maintained by the Manager
func (m *Manager) Shutdown() {
	for _, f := range m.files {
		os.Remove(f)
	}
}

func (m *Manager) Find(id string) (*os.File, error) {
	var file string
	for _, f := range m.files {
		if strings.Contains(f, id) {
			file = f
			break
		}
	}

	if file == "" {
		return nil, errors.New("Not Found")
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, errors.New("Error Opening file")
	}

	return f, nil
}

// NewFile created a new temporary file with format buffer*.tar.gzip
func NewFile(name string) (*os.File, error) {
	fname := name + "-*.tar.gzip"
	return ioutil.TempFile(dir, fname)
}

// WriteTarball walks the filepath starting at dir and writes the tarball into writer
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
