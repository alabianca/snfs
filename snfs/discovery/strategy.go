package discovery

import (
	"context"
	"net"
)

type Strategy interface {
	Register() error
	Shutdown()
	Lookup(ctx context.Context, instance string) ([]net.IP, error)
}

func MdnsStrategy(o ...Option) Strategy {
	return Service(o...)
}
