package discovery

import (
	"context"
	"net"
)

type Strategy interface {
	Register(instance string, port int) error
	Shutdown()
	Lookup(ctx context.Context, instance string) ([]net.IP, error)
}

func MdnsStrategy(o ...Option) Strategy {
	return Service(o...)
}
