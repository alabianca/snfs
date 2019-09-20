package peer

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/alabianca/snfs/snfs/fs"
)

type Manager struct {
	Addr       string
	Port       int
	storage    *fs.Manager
	httpServer *http.Server
}

func NewManager(storage *fs.Manager) *Manager {
	return &Manager{
		storage: storage,
	}
}

func (m *Manager) SetAddr(addr string, port int) {
	m.Addr = addr
	m.Port = port
}

func (m *Manager) REST() error {
	addr := net.JoinHostPort(m.Addr, strconv.Itoa(m.Port))

	m.httpServer = &http.Server{
		Addr:    addr,
		Handler: restAPIRoutes(m),
	}

	return m.httpServer.ListenAndServe()
}

func (m *Manager) Shutdown(ctx context.Context) error {
	if m.httpServer != nil {
		return m.httpServer.Shutdown(ctx)
	}
	return nil
}
