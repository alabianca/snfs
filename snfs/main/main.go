package main

import (
	"flag"
	"fmt"
	"github.com/alabianca/gokad"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/alabianca/snfs/snfs/kad"

	"github.com/alabianca/snfs/snfs/client"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/discovery"

	"github.com/alabianca/snfs/snfs/server"
)

const topLevelDomain = ".snfs.com"

var cport int
var dport int

func main() {
	myIP, err := util.MyIP("ipv4")
	// SNFS_CLIENT_CONNECTIVITY_PORT: Clis and other client applications connect to it
	cport = getPort("SNFS_CLIENT_CONNECTIVITY", 4200)
	// SNFS_DISCOVERY_PORT: This is the port is used to listen to discovery udp packets
	dport = getPort("SNFS_DISCOVERY", 5050)

	if err != nil {
		log.Fatal(err)
	}

	if err := util.SetEnv("SNFS_HOST", myIP.String()); err != nil {
		log.Fatal(err)
	}

	flag.Parse()
	done := make(chan os.Signal, 1)
	serverExit := make(chan error)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := server.New(dport, myIP.String())

	// initialize global queue channel
	server.InitQueue(server.NumServices)

	services := resolveServices(srv)
	for _, service := range services {
		srv.RegisterService(service)
	}

	go func() {
		serverExit <- srv.Run()
	}()

	// start the storage service immediately
	ss, _ := services[fs.ServiceName]
	startService(ss)

	// start the client connectivity service immediately
	cc, _ := services[client.ServiceName]
	startService(cc)

	select {
	case <-done:
		log.Println("Server Stopped...")
		srv.Shutdown()
	case <-serverExit:
		os.Exit(0)
	}

}

func startService(s server.Service) {
	req := server.ServiceRequest{
		Op:      server.OPStartService,
		Res:     make(chan server.ResponseCode, 1),
		Service: s,
	}

	server.QueueServiceRequest(req)
	<-req.Res
}

func resolveServices(s *server.Server) map[string]server.Service {
	rpc := kad.NewRPCManager(gokad.NewDHT(), s.Addr, s.Port)
	storage := fs.NewManager()
	dm := discovery.NewManager(discovery.MdnsStrategy(configureMDNS(s.Port, s.Addr, rpc.ID())))
	cc := client.NewConnectivityService(dm, storage, rpc)
	cc.SetAddr("", cport)

	services := map[string]server.Service{
		rpc.Name():     rpc,
		storage.Name(): storage,
		dm.Name():      dm,
		cc.Name():      cc,
	}

	return services

}

func configureMDNS(port int, address, id string) discovery.Option {
	return func(m *discovery.MdnsService) {
		text := []string{
			"Port:" + strconv.Itoa(port),
			"Address:" + address,
			"NodeID:" + id,
		}
		m.SetPort(port)
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

func getPort(prefix string, def int) int {
	p := os.Getenv(prefix + "_PORT")
	port, err := strconv.ParseInt(p, 10, 16)
	if err != nil {
		return def
	}

	return int(port)
}
