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

// MdnsService is a service that is discoverable in the local network
type MdnsService struct {
	instanceName string
	port         int
	ifaces       []net.Interface
	text         []string
	server       *zeroconf.Server
	service      string
	domain       string
}

// Option is a variadic configuration function to be passed to Server(option)
type Option func(s *MdnsService)

// Service creates a new service that is disoverable within the local network
func Service(o ...Option) MDNS {
	mdns := MdnsService{
		instanceName: "defaultInstanceName",
		port:         4200,
		ifaces:       nil,
		text:         nil,
		service:      ZeroConfService,
		domain:       ZeroConfDomain,
	}

	for _, opt := range o {
		opt(&mdns)
	}

	return &mdns
}

// Register registers the MdnsService in the local network
// at this point the service is disoverable under the "_snfs._tcp" service
func (mdns *MdnsService) Register() error {
	var err error
	mdns.server, err = zeroconf.Register(
		mdns.instanceName,
		mdns.service,
		mdns.domain,
		mdns.port,
		mdns.text,
		mdns.ifaces,
	)

	return err
}

// Shutdown closes all udp connections and closes the service
func (mdns *MdnsService) Shutdown() {
	if mdns.server != nil {
		mdns.server.Shutdown()
	}
}

// BrowseFor browses the local network for duration
func (mdns *MdnsService) BrowseFor(duration time.Duration) ([]*zeroconf.ServiceEntry, error) {
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

// Lookup finds a specific service instance
func (mdns *MdnsService) Lookup(ctx context.Context, instance string, entries chan *zeroconf.ServiceEntry) error {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return err
	}

	results := make(chan *zeroconf.ServiceEntry)
	go func(res chan *zeroconf.ServiceEntry) {
		for s := range res {
			if s.Instance == instance {
				entries <- s
			}
		}

		close(entries)
	}(results)

	resolver.Lookup(ctx, instance, ZeroConfService, ZeroConfDomain, results)

	return nil
}

func (mdns *MdnsService) Text() []string {
	return mdns.text
}

func (mdns *MdnsService) SetText(t []string) {
	mdns.text = t
}

func (mdns *MdnsService) Domain() string {
	return mdns.domain
}

func (mdns *MdnsService) SetDomain(d string) {
	mdns.domain = d
}

func (mdns *MdnsService) Interfaces() []net.Interface {
	return mdns.ifaces
}

func (mdns *MdnsService) SetInterfaces(ifaces []net.Interface) {
	mdns.ifaces = ifaces
}

func (mdns *MdnsService) Port() int {
	return mdns.port
}

func (mdns *MdnsService) SetPort(p int) {
	mdns.port = p
}

func (mdns *MdnsService) Instance() string {
	return mdns.instanceName
}

func (mdns *MdnsService) SetInstance(i string) {
	mdns.instanceName = i
}

func (mdns *MdnsService) Service() string {
	return mdns.service
}

func (mdns *MdnsService) SetService(s string) {
	mdns.service = s
}
