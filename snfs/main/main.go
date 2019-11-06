package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-homedir"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/discovery"

	"github.com/alabianca/snfs/snfs/server"
)

var dport = flag.Int("dp", 4000, "Port of the peer service, discoverable by peers")
var cport = flag.Int("cp", 4001, "Port of the client connectivity service. This port is used by local client applications")
var instance = flag.String("i", "default", "Instance Name")

const topLevelDomain = ".snfs.com"

func main() {
	myIP, err := util.MyIP("ipv4")

	if err != nil {
		log.Fatal(err)
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Mounting objects at ", home)

	flag.Parse()
	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	addr := ""

	server := server.Server{
		Port: 5050,
		Addr: myIP.String(),
	}

	server.InitializeDHT()

	// initialize file storage
	server.MountStorage(
		fs.NewManager(),
	)

	// set up discovery strategy
	server.SetDiscoveryManager(
		discovery.MdnsStrategy(configureMDNS(&server)),
	)

	if err := server.SetStoragePath(path.Join(home, "snfs")); err != nil {
		os.Exit(1)
	}

	// Client connectivity services like http/protobuf
	// TODO: Switch addr with real IP
	server.StartClientConnectivityService(addr, *cport)
	// Start the peer service. (service discoverable by other peers in local network)
	// TODO: Switch addr with real IP
	server.StartPeerService(addr, *dport)
	// Serve in separate go-routines
	clientConnectivityExited := serveHTTP(&server, server.ClientConnectivity)
	peerServiceExited := serveHTTP(&server, server.PeerService)

	// Wait for termination signal to attempt graceful shutdown
	<-done
	// Shutdown server gracefully
	log.Println("Server Stopped...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		// possibly more cleanup here
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown failed....")
	}

	<-clientConnectivityExited
	<-peerServiceExited
	log.Println("Server Shutdown properly")

}

func configureMDNS(s *server.Server) discovery.Option {
	return func(m *discovery.MdnsService) {
		m.SetPort(*dport)
		text := []string{
			"Port:" + strconv.Itoa(s.Port),
			"Address:" + s.Addr,
			"NodeID:" + fmt.Sprintf("%x", s.DHT.Table.ID.Bytes()),
		}
		m.SetText(text)
	}
}

func splitFromTopLevelDomain(instance string) (string, error) {
	if !strings.Contains(instance, topLevelDomain) {
		return "", fmt.Errorf("Instance does not contain top level domain")
	}
	split := strings.Split(instance, topLevelDomain)

	return split[0], nil
}

func serveHTTP(server *server.Server, service server.Rest) chan bool {
	done := make(chan bool)
	go func() {
		if err := server.HTTPListenAndServe(service); err != nil {
			done <- true
		}
	}()

	return done
}
