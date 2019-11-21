package server

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const NumServices = 5

const DiscoveryManager = "DiscoveryManager"
const ClientConnectivityService = "ConnectivityService"
const RPCManager = "RPCManager"
const StorageManager = "StorageManager"

var queue chan ServiceRequest
var onceQueue sync.Once
var onceInstance sync.Once
var serverInstance *Server
var mtx = sync.Mutex{}

type OP int
type ResponseCode int

const OPStartService = OP(1)
const OPStopService = OP(2)

const ResCodeServiceStarted = ResponseCode(200)
const ExitCodeSuccess = ResponseCode(0)
const ExitCodeError = ResponseCode(1)

type ServiceRequest struct {
	Op      OP
	Service Service
	Res     chan ResponseCode
}

func InitQueue(maxSize int) {
	onceQueue.Do(func() {
		queue = make(chan ServiceRequest, maxSize)
	})
}

func QueueServiceRequest(req ServiceRequest) {
	queue <- req
}

func ResolveService(token string) (Service, error) {
	mtx.Lock()
	defer mtx.Unlock()
	if serverInstance == nil {
		return nil, errors.New(ErrServiceNotResolved)
	}

	s, ok := serverInstance.services[token]
	if !ok {
		return nil, errors.New(ErrServiceNotResolved)
	}

	return s.service, nil
}

// Errors
const ErrServerNotSet = "Server Not Set"
const ErrServiceNotResolved = "Service Not Resolved"

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

	services    map[string]*serviceEntry
	lock        sync.Mutex
	stopAll     chan bool
	exit        chan error
	exitService chan string
}

func New(port int, host string) *Server {
	onceInstance.Do(func() {
		serverInstance = &Server{
			Port:        port,
			Addr:        host,
			lock:        sync.Mutex{},
			services:    make(map[string]*serviceEntry),
			stopAll:     make(chan bool),
			exit:        make(chan error),
			exitService: make(chan string),
		}
	})

	return serverInstance
}

func (s *Server) RegisterService(service Service) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.services[service.Name()] = &serviceEntry{
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
	pendingRequests := make([]ServiceRequest, 0)
	var nextRun <-chan time.Time
	for {

		var next ServiceRequest
		if len(pendingRequests) > 0 {
			next = pendingRequests[0]
			nextRun = time.After(time.Millisecond * 200)
		}

		select {
		case req := <-queue:
			pendingRequests = append(pendingRequests, req)

		case <-nextRun:
			pendingRequests = pendingRequests[1:]
			go func(req ServiceRequest) {
				log.Printf("Handling Request %d for %s\n", req.Op, req.Service.Name())
				s.handleRequest(&req)
			}(next)

		case <-s.stopAll:
			log.Printf("Stopping All pendingServices\n")
			for _, j := range s.services {
				log.Printf("Stopping %s %v\n", j.service.Name(), j.started)
				if j.started {
					j.service.Shutdown()
					s.exitService <- fmt.Sprintf("Stopped Service (%s)\n", j.service.Name())
				}
			}

			close(s.exitService)

			log.Println("All Services Stopped")

			s.exit <- nil
		}
	}
}

func (s *Server) handleRequest(req *ServiceRequest) {
	switch req.Op {
	case OPStartService:
		s.startService(req)
	case OPStopService:
		s.stopService(req)
	}
}

func (s *Server) startService(req *ServiceRequest) {
	entry, _ := s.services[req.Service.Name()]
	if entry.started {
		return
	}
	entry.started = true
	go func() {
		entry.service.Run()
		log.Printf("%s mainloop exited\n", entry.service.Name())
	}()

	req.Res <- ResCodeServiceStarted

}

func (s *Server) stopService(req *ServiceRequest) {
	entry, _ := s.services[req.Service.Name()]
	entry.started = false

	if err := entry.service.Shutdown(); err != nil {
		req.Res <- ExitCodeError
	}
	req.Res <- ExitCodeSuccess
}
