package discovery

import (
	"context"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	// ZeroConfService is the service name of all snfs services
	ZeroConfService = "_snfs._tcp"
	// ZeroConfDomain '.local'
	ZeroConfDomain = ".local"
)

// MdnsServer is a service that is discoverable in the local network
type MdnsServer struct {
	InstanceName string
	Port         int
	Ifaces       []net.Interface
	Text         []string
	server       *zeroconf.Server
	service      string
	domain       string
}

// Option is a variadic configuration function to be passed to Server(option)
type Option func(s *MdnsServer)

// Server creates a new service that is disoverable within the local network
func Server(o ...Option) *MdnsServer {
	mdns := MdnsServer{
		InstanceName: "defaultInstanceName",
		Port:         4200,
		Ifaces:       nil,
		Text:         nil,
		service:      ZeroConfService,
		domain:       ZeroConfDomain,
	}

	for _, opt := range o {
		opt(&mdns)
	}

	return &mdns
}

// Register registers the MdnsServer in the local network
// at this point the service is disoverable under the "_snfs._tcp" service
func (mdns *MdnsServer) Register() error {
	var err error
	mdns.server, err = zeroconf.Register(
		mdns.InstanceName,
		mdns.service,
		mdns.domain,
		mdns.Port,
		mdns.Text,
		mdns.Ifaces,
	)

	return err
}

// Shutdown closes all udp connections and closes the service
func (mdns *MdnsServer) Shutdown() {
	if mdns.server != nil {
		mdns.server.Shutdown()
	}
}

// BrowseFor browses the local network for duration
func (mdns *MdnsServer) BrowseFor(duration time.Duration) ([]*zeroconf.ServiceEntry, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	out := make([]*zeroconf.ServiceEntry, 0)
	go func(res <-chan *zeroconf.ServiceEntry) {
		for entry := range res {
			out = append(out, entry)
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), duration)

	defer cancel()

	err = resolver.Browse(ctx, ZeroConfService, ZeroConfDomain, entries)
	if err != nil {
		return nil, err
	}

	<-ctx.Done()

	return out, nil

}
