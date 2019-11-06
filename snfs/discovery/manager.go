package discovery

import (
	"context"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

type Manager struct {
	ResolveTimeout time.Duration
	BrowseTimeout  time.Duration
	mdns           *MdnsService
}

func NewManager(mdns *MdnsService) *Manager {
	m := Manager{
		ResolveTimeout: time.Second * 5,
		BrowseTimeout:  time.Second * 5,
		//strategy:       s,
		mdns: mdns,
	}

	return &m
}

func (m *Manager) Register(instance string) error {
	return m.mdns.Register(instance)
}

func (m *Manager) UnRegister() {
	m.mdns.isStarted = false
	m.mdns.Shutdown()
}

func (m *Manager) Resolve(instance string) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.ResolveTimeout)
	defer cancel()

	return m.mdns.Lookup(ctx, instance)
}

func (m *Manager) Browse() ([]*zeroconf.ServiceEntry, error) {
	return m.mdns.BrowseFor(m.BrowseTimeout)
}

func (m *Manager) MDNSStarted() bool {
	return m.mdns.IsStarted()
}
