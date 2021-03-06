package fs

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/alabianca/snfs/util"

	"github.com/mitchellh/go-homedir"
)

const ServiceName = "StorageManager"

// Manager maintains a list of files that are currently
// shared in the local network
type Manager struct {
	root    string
	objects map[string]*object
	fileServer *server
	id      []byte
}

// NewManager returns a Manager with zero files
func NewManager() *Manager {
	m := &Manager{
		objects: make(map[string]*object),
		id:      make([]byte, 20),
	}

	util.RandomID(m.id)

	return m
}

// GetRoot returns the manager's root storage path
func (m *Manager) GetRoot() string {
	return m.root
}

func (m *Manager) SetRoot(dir string) {
	root := "objects"
	m.root = path.Join(dir, root)

}

// CreateRootDir creates the root dir if it does not yet exist
// calling CreateRootDir before calling SetRoot results in a "Root Unset" error
func (m *Manager) CreateRootDir() error {
	if m.root == "" {
		return errors.New("Root Unset")
	}

	if _, err := os.Stat(m.root); os.IsNotExist(err) {
		if err := os.MkdirAll(m.root, 0700); err != nil {
			return err
		}
	}

	return nil
}

// AddObject adds a file object to manager's memory
// if the object with specified name already exists, the manager attempts to delete it
// if there is an error it is of type PathError as a result of a failed deletion
func (m *Manager) AddObject(name, hash string, size int64) error {
	_, ok := m.objects[name]
	if ok {
		if err := m.delete(name); err != nil {
			return err
		}
	}

	m.objects[name] = &object{
		name: name,
		hash: hash,
		size: size,
	}

	return nil
}

func (m *Manager) GetObjectPath(name string) (string, error) {
	obj, ok := m.objects[name]
	if !ok {
		return "", errors.New("Object Not Found")
	}

	return path.Join(m.root, obj.hash), nil
}

func (m *Manager) delete(name string) error {
	if err := os.Remove(path.Join(m.root, name)); err != nil {
		return err
	}

	delete(m.objects, name)

	return nil
}

// Service

func (m *Manager) Run() error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	m.SetRoot(path.Join(home, "snfs"))
	if err := m.CreateRootDir(); err != nil {
		return err
	}

	port, err := strconv.ParseInt(os.Getenv("SNFS_FS_PORT"), 10, 16)
	if err != nil {
		return err
	}
	m.fileServer = &server{
		addr: os.Getenv("SNFS_HOST"),
		port: int(port),
	}

	return m.fileServer.listen(m)
}

func (m *Manager) ID() string {
	return fmt.Sprintf("%x", m.id)
}

func (m *Manager) Shutdown() error {
	for k, _ := range m.objects {
		m.delete(k)
	}

	return nil
}

func (m *Manager) Name() string {
	return ServiceName
}

// NewFile created a new temporary file with format buffer*.tar.gzip
func NewFile(dir, name string) (*os.File, error) {
	fname := path.Join(dir, name)
	return os.Create(fname)

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
