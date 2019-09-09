package client

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/alabianca/snfs/snfs/discovery"
)

type ConnectivityService struct {
	Addr       string
	Port       int
	discovery  *discovery.Manager
	httpServer *http.Server
}

func NewConnectivityService(dManager *discovery.Manager) *ConnectivityService {
	return &ConnectivityService{
		discovery: dManager,
	}
}

func (c *ConnectivityService) SetAddr(addr string, port int) {
	c.Addr = addr
	c.Port = port
}

func (c *ConnectivityService) REST() error {
	addr := net.JoinHostPort(c.Addr, strconv.Itoa(c.Port))

	c.httpServer = &http.Server{
		Addr:    addr,
		Handler: restAPIRoutes(c),
	}

	return c.httpServer.ListenAndServe()
}

func (c *ConnectivityService) Shutdown(ctx context.Context) error {
	if c.httpServer != nil {
		return c.httpServer.Shutdown(ctx)
	}
	return nil
}
