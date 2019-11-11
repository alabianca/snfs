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

	// server := server.Server{
	// 	Port: 5050,
	// 	Addr: myIP.String(),
	// }
	server := server.New(5050, myIP.String())

	services := resolveServices(server)

	go func() {
		serverExit <- server.Run()
	}()

	// start the client connectivity service immediately
	cc, _ := services[client.ServiceName]
	server.StartJob(cc)

	select {
	case <-done:
		log.Println("Server Stopped...")
		server.Shutdown()
	case <-serverExit:
		os.Exit(0)
	}

}

func resolveServices(s *server.Server) map[string]server.Job {
	rpc := s.InitializeDHT()
	storage := s.MountStorage(fs.NewManager())
	dm := s.SetDiscoveryManager(discovery.MdnsStrategy(configureMDNS(s)))
	cc := s.StartClientConnectivityService(*cport)

	services := map[string]server.Job{
		rpc.Name():     rpc,
		storage.Name(): storage,
		dm.Name():      dm,
		cc.Name():      cc,
	}

	return services

}

func configureMDNS(s *server.Server) discovery.Option {
	return func(m *discovery.MdnsService) {
		m.SetPort(*dport)
		text := []string{
			"Port:" + strconv.Itoa(s.Port),
			"Address:" + s.Addr,
			"NodeID:" + s.GetOwnID(),
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

// func serveHTTP(server *server.Server, service server.Rest) chan bool {
// 	done := make(chan bool)
// 	go func() {
// 		if err := server.HTTPListenAndServe(service); err != nil {
// 			done <- true
// 		}
// 	}()

// 	return done
// }
