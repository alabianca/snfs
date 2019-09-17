package server

import (
	"context"
	"fmt"
	"log"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/client"

	"github.com/alabianca/snfs/snfs/discovery"
)

type Server struct {
	Port               int
	Addr               string
	DiscoveryManager   *discovery.Manager
	ClientConnectivity *client.ConnectivityService
	FsManager          *fs.Manager
}

func (s *Server) SetDiscoveryManager(strategy discovery.Strategy) {
	s.DiscoveryManager = discovery.NewManager(strategy)
}

func (s *Server) StartClientConnectivityService() {
	s.ClientConnectivity = client.NewConnectivityService(s.DiscoveryManager, s.FsManager)
	s.ClientConnectivity.SetAddr(s.Addr, s.Port)

}

func (s *Server) HTTPListenAndServe() error {
	if s.ClientConnectivity == nil {
		return fmt.Errorf("Client Connectivity Service not started")
	}

	return s.ClientConnectivity.REST()

}

func (s *Server) Shutdown(ctx context.Context) error {
	// cleanup any temporary files
	defer func() {
		log.Println("Removing Temporary Files")
		if err := s.FsManager.Cleanup(); err != nil {
			log.Fatal(err)
		}
	}()

	// stop mdns
	s.DiscoveryManager.UnRegister()
	// stop accepting connections
	return s.ClientConnectivity.Shutdown(ctx)
}
