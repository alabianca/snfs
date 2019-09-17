package fs

import (
	"fmt"
	"io/ioutil"
	"os"
)

const dir = "/Users/alexanderlabianca/go/src/github.com/alabianca/snfs"

type Manager struct {
	tempFile *os.File
}

func NewManager() *Manager {
	return &Manager{}

}

func (m *Manager) MakeTempFile() error {
	if m.tempFile != nil {
		if err := os.Remove(m.tempFile.Name()); err != nil {
			return err
		}
	}

	file, err := ioutil.TempFile(dir, "buffer")
	if err != nil {
		return err
	}

	m.tempFile = file

	return nil
}

func (m *Manager) Write(p []byte) (n int, err error) {
	if m.tempFile == nil {
		return 0, fmt.Errorf("File not set")
	}
	return m.tempFile.Write(p)
}

func (m *Manager) Read(p []byte) (n int, err error) {
	return m.tempFile.Read(p)
}

func (m *Manager) Cleanup() error {
	defer os.Remove(m.tempFile.Name())
	if err := m.tempFile.Close(); err != nil {
		return err
	}

	return nil
}
