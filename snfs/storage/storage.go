package storage

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/alabianca/snfs/snfs"
	"github.com/alabianca/snfs/snfs/storage/memory"
	"github.com/alabianca/snfs/snfs/storage/writer"
	"io"
	"net"
	"net/http"
	"os"
	"path"
)

type Config struct {
	Root        string
	Addr        string
	Port        int
	Storage     StorageMechanismFunc
}

type StorageMechanismFunc func() StorageMechanism

func Memory() StorageMechanism {
	return memory.New()
}

const ErrObjectNotStored = "object not stored"

type StorageMechanism interface {
	Store(key string, object interface{}) bool
	Get(key string, val interface{}) bool
	Clear()
}

type StorageService interface {
	snfs.Storage
}

type storageService struct {
	network      snfs.NetworkClient
	localStorage StorageMechanism
	root         string
	addr         string
	port         int
}

func New(node snfs.NetworkClient, config Config) StorageService {
	s := &storageService{
		network: node,
		root:    config.Root,
		addr:    config.Addr,
		port:    config.Port,
		localStorage: config.Storage(),
	}

	return s
}

func (s *storageService) Shutdown() error {
	s.localStorage.Clear()
	return nil
}

func (s *storageService) Get(fileHash string) ([]byte, error) {
	addr, err := s.network.Resolve(fileHash)
	if err != nil {
		return nil, err
	}

	return s.requestFile(addr, fileHash)
}

func (s *storageService) Store(fileName string, reader io.Reader) (int, error) {
	// create a new file
	file, err := newFile(s.root, fileName)
	if err != nil {
		return 0, err
	}

	// write the file to 'file' and hash it
	storageWriter := writer.New(file)
	n, err := io.Copy(storageWriter, reader)
	if err != nil {
		return 0, err
	}

	// create an object and store it locally
	hash := fmt.Sprintf("%x", storageWriter.Sum(nil))
	ok := s.localStorage.Store(fileName, &Object{
		Name: fileName,
		Hash: hash,
		Size: n,
		Path: path.Join(s.root, fileName),
	})

	if !ok {
		return 0, errors.New(ErrObjectNotStored)
	}

	return s.storeInNetwork(hash)
}

func (s *storageService) storeInNetwork(hashed string) (int, error) {
	return s.network.Store(
		hashed,
		net.ParseIP(s.addr),
		s.port,
	)
}

func (s *storageService) requestFile(addr net.Addr, fileHash string) ([]byte, error) {
	url := "http://" + addr.String() + "/v1/object/" + fileHash
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, res.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func newFile(dir, name string) (*os.File, error) {
	fname := path.Join(dir, name)
	return os.Create(fname)
}
