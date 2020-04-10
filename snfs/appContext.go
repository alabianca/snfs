package snfs

import (
	"io"
	"net"
)


type AppContext struct {
	Storage Storage
}

type Storage interface {
	Store(filename string, reader io.Reader) (int, error)
	Get(fileHash string) ([]byte, error)
}

type Resolver interface {
	Resolve(fileHash string) (net.Addr, error)
}

type RPC interface {
	Store(key string, ip net.IP, port int) (int, error)
}

type NetworkClient interface {
	Resolver
	RPC
}

type NetworkHost interface {
	Bootstrap(port int, ip string) error
	Shutdown()
	Listen() error
	ID() string
}

type Network interface {
	NetworkClient
	NetworkHost
}