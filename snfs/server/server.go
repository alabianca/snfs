package server

import (
	"context"
	"errors"
	"log"

	"github.com/alabianca/snfs/snfs/kadnet"
	"github.com/alabianca/snfs/snfs/peer"

	"github.com/alabianca/snfs/snfs/client"
	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/discovery"
)

type Server struct {
	Port               int
	Addr               string
	DiscoveryManager   *discovery.Manager
	ClientConnectivity *client.ConnectivityService
	PeerService        *peer.Manager
	Storage            *fs.Manager
	RpcManager         kadnet.RPC
}

func (s *Server) InitializeDHT() {
	log.Printf("Initializing DHT at %s -> %d\n", s.Addr, s.Port)
	s.RpcManager = kadnet.NewRPCManager(s.Addr, s.Port)
}

func (s *Server) MountStorage(storage *fs.Manager) {
	s.Storage = storage
}

func (s *Server) SetStoragePath(path string) error {
	if s.Storage == nil {
		return errors.New("Storage Manager Not Set")
	}

	s.Storage.SetRoot(path)
	if err := s.Storage.CreateRootDir(); err != nil {
		return err
	}

	return nil
}

func (s *Server) SetDiscoveryManager(mdns *discovery.MdnsService) {
	s.DiscoveryManager = discovery.NewManager(mdns)
}

func (s *Server) StartClientConnectivityService(addr string, port int) {
	s.ClientConnectivity = client.NewConnectivityService(s.DiscoveryManager, s.Storage, s.RpcManager)
	s.ClientConnectivity.SetAddr(addr, port)

}

func (s *Server) StartPeerService(addr string, port int) {
	s.PeerService = peer.NewManager(s.Storage)
	s.PeerService.SetAddr(addr, port)
}

func (s *Server) HTTPListenAndServe(service Rest) error {
	if service == nil {
		return errors.New("nil service provided")
	}

	return service.REST()
}

func (s *Server) GetOwnID() []byte {
	return s.RpcManager.GetID()
}

func (s *Server) Shutdown(ctx context.Context) error {
	defer func() {
		s.Storage.Shutdown()
	}()

	// stop mdns
	s.DiscoveryManager.UnRegister()
	// stop accepting connections
	s.ClientConnectivity.Shutdown(ctx)
	return s.PeerService.Shutdown(ctx)
}
