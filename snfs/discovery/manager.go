package discovery

import (
	"context"
	"net"
	"time"
)

type Manager struct {
	ResolveTimeout time.Duration
	strategy       Strategy
}

func NewManager(s Strategy) *Manager {
	m := Manager{
		ResolveTimeout: time.Second * 5,
		strategy:       s,
	}

	return &m
}

func (m *Manager) Register(instance string) error {
	return m.strategy.Register(instance)
}

func (m *Manager) UnRegister() {
	m.strategy.Shutdown()
}

func (m *Manager) Resolve(instance string) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.ResolveTimeout)
	defer cancel()

	return m.strategy.Lookup(ctx, instance)
}
