package server

import (
	"fmt"
	"log"
	"sync"
)

const NumServices = 5

var queue chan Service
var once sync.Once

func InitQueue(maxSize int) {
	once.Do(func() {
		queue = make(chan Service, maxSize)
	})
}

func QueueService(s Service) {
	queue <- s
}

// Errors
const ErrServerNotSet = "Server Not Set"

type serviceEntry struct {
	started bool
	service Service
}

type Service interface {
	Run() error
	Shutdown() error
	ID() string
	Name() string
}

type Server struct {
	Port int
	Addr string

	services     map[string]serviceEntry
	lock         sync.Mutex
	startService chan Service
	stopService  chan Service
	stopAll      chan bool
	exit         chan error
	exitService  chan string
}

func New(port int, host string) *Server {
	serverInstace := &Server{
		Port:         port,
		Addr:         host,
		lock:         sync.Mutex{},
		services:     make(map[string]serviceEntry),
		startService: make(chan Service, NumServices),
		stopService:  make(chan Service, 1),
		stopAll:      make(chan bool),
		exit:         make(chan error),
		exitService:  make(chan string),
	}

	return serverInstace
}

func (s *Server) StartService(service Service) {
	log.Printf("Starting Service %s (%s)\n", service.ID(), service.Name())
	s.startService <- service
}

func (s *Server) RegisterService(service Service) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.services[service.Name()] = serviceEntry{
		started: false,
		service: service,
	}
}

func (s *Server) ResolveService(token string) Service {
	entry, ok := s.services[token]
	if !ok {
		return nil
	}

	return entry.service
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
	services := make([]Service, 0)
	for {
		select {
		case service := <-queue:
			s.startService <- service
		case service := <-s.startService:
			services = append(services, service)
			go func(s Service) {
				log.Printf("Running %s (%s)\n", service.ID(), service.Name())
				s.Run()
			}(service)
		case service := <-s.stopService:
			for _, j := range services {
				if j.ID() == service.ID() {
					j.Shutdown()
					log.Printf("Stopped Service %s (%s)\n", j.ID(), j.Name())
				}
			}

		case <-s.stopAll:
			log.Printf("Stopping All Services (%d)\n", len(services))
			for _, j := range services {
				j.Shutdown()
				s.exitService <- fmt.Sprintf("Stopped Service %s (%s)\n", j.ID(), j.Name())
			}

			close(s.exitService)

			log.Println("All Services Stopped")

			s.exit <- nil
		}
	}
}
