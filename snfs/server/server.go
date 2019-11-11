package server

import (
	"fmt"
	"log"

	"github.com/alabianca/snfs/snfs/kadnet"

	"github.com/alabianca/snfs/snfs/client"
	"github.com/alabianca/snfs/snfs/fs"

	"github.com/alabianca/snfs/snfs/discovery"
)

const numJobs = 5

type Job interface {
	Run() error
	Shutdown() error
	ID() string
	Name() string
}

type startJob struct {
	job Job
	err error
}

type Server struct {
	Port               int
	Addr               string
	DiscoveryManager   *discovery.Manager
	ClientConnectivity *client.ConnectivityService
	Storage            *fs.Manager
	RPCManager         kadnet.RPCManager
	// channels
	startJob    chan Job
	stopJob     chan Job
	stopAll     chan bool
	exit        chan error
	exitService chan string
}

func New(port int, host string) *Server {
	return &Server{
		Port:        port,
		Addr:        host,
		startJob:    make(chan Job, numJobs),
		stopJob:     make(chan Job, 1),
		stopAll:     make(chan bool),
		exit:        make(chan error),
		exitService: make(chan string),
	}
}

func (s *Server) StartJob(job Job) {
	log.Printf("Starting Job %s (%s)\n", job.ID(), job.Name())
	s.startJob <- job
}

func (s *Server) SetRPCManager() Job {
	log.Printf("Initializing DHT at %s -> %d\n", s.Addr, s.Port)
	s.RPCManager = kadnet.NewRPCManager(s.Addr, s.Port)
	return s.RPCManager
}

func (s *Server) SetStorageManager(storage *fs.Manager) Job {
	s.Storage = storage
	return s.Storage
}

func (s *Server) SetDiscoveryManager(mdns *discovery.MdnsService) Job {
	s.DiscoveryManager = discovery.NewManager(mdns)
	return s.DiscoveryManager
}

func (s *Server) SetClientConnectivityService(port int) Job {
	s.ClientConnectivity = client.NewConnectivityService(s.DiscoveryManager, s.Storage, s.RPCManager)
	s.ClientConnectivity.SetAddr("", port)
	return s.ClientConnectivity
}

func (s *Server) GetOwnID() string {
	return s.RPCManager.ID()
}

func (s *Server) Run() error {
	log.Printf("Running Server at %s:%d\n", s.Addr, s.Port)

	go s.mainLoop()

	err := <-s.exit
	log.Println("Exiting Main Loop")
	return err
}

func (s *Server) Shutdown() {
	s.stopAll <- true
	for s := range s.exitService {
		log.Println(s)
	}
}

func (s *Server) mainLoop() {
	jobs := make([]Job, 0)
	for {
		select {
		case job := <-s.startJob:
			jobs = append(jobs, job)
			go func() {
				log.Printf("Running %s (%s)\n", job.ID(), job.Name())
				job.Run()
			}()
		case job := <-s.stopJob:
			for _, j := range jobs {
				if j.ID() == job.ID() {
					j.Shutdown()
					log.Printf("Stopped Service %s (%s)\n", j.ID(), j.Name())
				}
			}

		case <-s.stopAll:
			log.Printf("Stopping All Services (%d)\n", len(jobs))
			for _, j := range jobs {
				j.Shutdown()
				s.exitService <- fmt.Sprintf("Stopped Service %s (%s)\n", j.ID(), j.Name())
			}

			close(s.exitService)

			log.Println("All Services Stopped")

			s.exit <- nil
		}
	}
}
