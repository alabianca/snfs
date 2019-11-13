package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/alabianca/snfs/snfs/kadnet"

	"github.com/alabianca/snfs/snfs/client"

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

	flag.Parse()
	done := make(chan os.Signal, 1)
	serverExit := make(chan error)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := server.New(5050, myIP.String())

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
	log.Printf("Started Service %s\n", s.Name())
}

func resolveServices(s *server.Server) map[string]server.Service {
	rpc := kadnet.NewRPCManager(s.Addr, s.Port)
	storage := fs.NewManager()
	dm := discovery.NewManager(discovery.MdnsStrategy(configureMDNS(s.Port, s.Addr, rpc.ID())))
	cc := client.NewConnectivityService(dm, storage, rpc)
	cc.SetAddr("", *cport)

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
		m.SetPort(*dport)
		text := []string{
			"Port:" + strconv.Itoa(port),
			"Address:" + address,
			"NodeID:" + id,
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
