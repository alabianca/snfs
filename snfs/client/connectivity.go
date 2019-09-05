package client

import (
	"net"
	"net/http"
	"strconv"

	"github.com/alabianca/snfs/snfs/discovery"
)

type ConnectivityService struct {
	Addr      string
	Port      int
	discovery *discovery.Manager
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

	return http.ListenAndServe(
		addr,
		restAPIRoutes(c),
	)
}
