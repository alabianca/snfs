package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/alabianca/snfs/util"

	"github.com/grandcat/zeroconf"
)

const ServiceName = "DiscoveryManager"

// Errors
const ErrInstanceNotSet = "Instance Has Not Been Set"

type Manager struct {
	ResolveTimeout time.Duration
	BrowseTimeout  time.Duration
	mdns           *MdnsService
	instance       string
	id             []byte
}

func NewManager(mdns *MdnsService) *Manager {
	m := Manager{
		ResolveTimeout: time.Second * 5,
		BrowseTimeout:  time.Second * 5,
		mdns:           mdns,
		id:             make([]byte, 20),
	}

	util.RandomID(m.id)

	return &m
}

func (m *Manager) SetInstance(instance string) {
	m.instance = instance
}

func (m *Manager) Register(instance string) error {
	return m.mdns.Register(instance)
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

// JOB

func (m *Manager) Run() error {
	if m.instance == "" {
		return errors.New(ErrInstanceNotSet)
	}

	if err := m.Register(m.instance); err != nil {
		return err
	}

	return nil
}

func (m *Manager) ID() string {
	return fmt.Sprintf("%x", m.id)
}

func (m *Manager) Shutdown() error {
	m.mdns.isStarted = false
	m.mdns.Shutdown()
	return nil
}

func (m *Manager) Name() string {
	return ServiceName
}
