package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/discovery"

	"github.com/alabianca/snfs/snfs/server"
)

var port = flag.Int("p", 4000, "Port of the server")
var instance = flag.String("i", "default", "Instance Name")

const topLevelDomain = ".snfs.com"

func main() {
	flag.Parse()
	done := make(chan os.Signal, 1)
	httpClosed := make(chan bool)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	addr := ""

	server := server.Server{
		Port: *port,
		Addr: addr,
	}

	// initialize file storage
	server.MountStorage(
		fs.NewManager(),
	)

	// set up discovery strategy
	server.SetDiscoveryManager(
		discovery.MdnsStrategy(mdnsConfig),
	)

	// Client connectivity services like http/protobuf
	server.StartClientConnectivityService()

	// Serve in separate thread
	go func() {
		if err := server.HTTPListenAndServe(); err != nil {
			log.Println(err)
			httpClosed <- true
		}
	}()

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

	<-httpClosed
	log.Println("Server Shutdown properly")

}

func mdnsConfig(m *discovery.MdnsService) {
	instanceName, _ := splitFromTopLevelDomain(*instance)
	m.SetPort(*port)
	m.SetInstance(instanceName)

}

func splitFromTopLevelDomain(instance string) (string, error) {
	if !strings.Contains(instance, topLevelDomain) {
		return "", fmt.Errorf("Instance does not contain top level domain")
	}
	split := strings.Split(instance, topLevelDomain)

	return split[0], nil
}
