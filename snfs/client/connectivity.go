package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/discovery"
	"github.com/alabianca/snfs/snfs/kadnet"
)

const ServiceName = "ClientConnectivityService"

type ConnectivityService struct {
	Addr       string
	Port       int
	discovery  *discovery.Manager
	httpServer *http.Server
	storage    *fs.Manager
	rpc        *kadnet.RpcManager
	id         []byte
	name       string
}

func NewConnectivityService(dManager *discovery.Manager, storage *fs.Manager, rpc *kadnet.RpcManager) *ConnectivityService {

	c := &ConnectivityService{
		discovery: dManager,
		storage:   storage,
		rpc:       rpc,
		id:        make([]byte, 20),
	}

	util.RandomID(c.id)

	return c
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

// JOB

func (c *ConnectivityService) Run() error {
	return c.REST()
}

func (c *ConnectivityService) ID() string {
	return fmt.Sprintf("%x", c.id)
}

func (c *ConnectivityService) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if c.httpServer != nil {
		return c.httpServer.Shutdown(ctx)
	}
	return nil
}

func (c *ConnectivityService) Name() string {
	return ServiceName
}
