package discovery

import (
	"context"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

type Strategy interface {
	Register(instance string) error
	Shutdown()
	Lookup(ctx context.Context, instance string) ([]net.IP, error)
	BrowseFor(duration time.Duration) ([]*zeroconf.ServiceEntry, error)
}

func MdnsStrategy(o ...Option) Strategy {
	return Service(o...)
}
