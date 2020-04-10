package main

import (
	"flag"
	"github.com/alabianca/snfs/snfs"
	"github.com/alabianca/snfs/snfs/http"
	"github.com/alabianca/snfs/snfs/http/chi"
	"github.com/alabianca/snfs/snfs/kademlia"
	"github.com/alabianca/snfs/snfs/storage"
	"github.com/alabianca/snfs/util"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	cport   *int
	dport   *int
	rootDir *string
	id      *int64
)

const (
	tag = "main"
)

func init() {
	cport = flag.Int("cport", 4200, "Gateway port. Clis and other applications may connect to it")
	dport = flag.Int("dport", 5050, "This port listens to kademlia packets")
	rootDir = flag.String("root", ".", "The root directory of this node")
	id = flag.Int64("id", 0, "The local id of the process started by the watchdog")
}

func main() {
	flag.Parse()
	// add a global event logger
	snfs.EventLogger = snfs.NewLogger(os.Stdout)
	snfs.EventLogger.Start()
	// find my private ipv4 address
	myIP, err := util.MyIP("ipv4")
	if err != nil {
		panic(err)
	}

	// start the kademlia node
	network := kademlia.New(myIP.String(), *dport)

	storageService := storage.New(network, storage.Config{
		Root:    *rootDir,
		Addr:    myIP.String(),
		Port:    *dport,
		Storage: storage.Memory,
	})

	appContext := snfs.AppContext{
		Storage: storageService,
	}

	handler := http.App(&appContext, chi.Routes)

	server := snfs.Server{
		Handler: handler,
		Host:    myIP.String(),
		Port:    *cport,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, os.Kill)

	go func() {
		network.Listen()
	}()

	exit := make(chan struct{})
	go func() {
		<-done
		close(exit)
	}()

	snfs.EventLogger.NewEvent(tag, "NodeStarted", &snfs.NodeConfiguration{
		ID:        *id,
		ProcessId: os.Getpid(),
		NodeID:    network.ID(),
		Cport: *cport,
		Dport: *dport,
	})

	server.Run(exit)
	snfs.EventLogger.NewEvent(tag, "NodeExited", &snfs.NodeConfiguration{
		ID:        *id,
		NodeID:    network.ID(),
		ProcessId: os.Getpid(),
		Dport: *dport,
		Cport: *cport,
	})


	// once the server returns we have to also shutdown our network node
	network.Shutdown()
	snfs.EventLogger.Exit()
	time.Sleep(time.Second * 2)
}
