package server

import (
	"fmt"

	"github.com/alabianca/snfs/snfs/client"

	"github.com/alabianca/snfs/snfs/discovery"
)

type Server struct {
	Port               int
	Addr               string
	DiscoveryManager   *discovery.Manager
	ClientConnectivity *client.ConnectivityService
}

func (s *Server) SetDiscoveryManager(strategy discovery.Strategy) {
	s.DiscoveryManager = discovery.NewManager(strategy)
}

func (s *Server) StartClientConnectivityService() {
	s.ClientConnectivity = client.NewConnectivityService(s.DiscoveryManager)
	s.ClientConnectivity.SetAddr(s.Addr, s.Port)

}

func (s *Server) HTTPListenAndServe() error {
	if s.ClientConnectivity == nil {
		return fmt.Errorf("Client Connectivity Service not started")
	}

	return s.ClientConnectivity.REST()

}
