package main

import (
	"flag"
	"github.com/alabianca/gokad"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/alabianca/snfs/snfs/kad"

	"github.com/alabianca/snfs/snfs/client"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/discovery"

	"github.com/alabianca/snfs/snfs/server"
)

const topLevelDomain = ".snfs.com"

var cport *int
var dport *int
var fport *int
func init() {
	cport = flag.Int("cport", 4200, "Gateway port. Clis and other applications may connect to it")
	dport = flag.Int("dport", 5050, "This port listens to mdns and kademlia packets")
	fport = flag.Int("fport", 7000, "The file server lives at this port")
}

func main() {
	myIP, err := util.MyIP("ipv4")
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()
	server.GetGlobals().Set("SNFS_CLIENT_CONNECTIVITY", strconv.Itoa(*cport))
	server.GetGlobals().Set("SNFS_DISCOVERY_PORT", strconv.Itoa(*dport))
	server.GetGlobals().Set("SNFS_FS_PORT", strconv.Itoa(*fport))
	server.GetGlobals().Set("SNFS_HOST", myIP.String())


	done := make(chan os.Signal, 1)
	serverExit := make(chan error)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// create a server instance
	srv := server.New(*dport, myIP.String())
	// initialize global queue channel
	server.InitQueue(server.NumServices)

	services := registerServices(srv)
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

func registerServices(server *server.Server) map[string]server.Service {
	services := resolveServices(server)
	for _, s := range services {
		server.RegisterService(s)
	}

	return services
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
		text := []string{
			"Port:" + strconv.Itoa(port),
			"Address:" + address,
			"NodeID:" + id,
		}
		m.SetPort(port)
		m.SetText(text)
	}
}
