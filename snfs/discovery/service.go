package discovery

import (
	"context"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

type MDNS interface {
	BrowseFor(duration time.Duration) ([]*zeroconf.ServiceEntry, error)
	Lookup(ctx context.Context, instance string, entries chan *zeroconf.ServiceEntry) error
	Text() []string
	Domain() string
	Service() string
	Interfaces() []net.Interface
	Port() int
	Instance() string
	Register() error
	Shutdown()
}
