package server

import (
	"context"
	"github.com/alabianca/snfs/snfsd"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	Watchdog snfsd.Watchdog
	Handler  snfsd.Handler
	Host     string
	Port     int
	server   *http.Server
}

func (s *Server) Run(exit chan struct{}) error {
	var wg sync.WaitGroup
	wg.Add(2)

	var err error

	go func() {
		s.Watchdog.Watch()
		wg.Done()
	}()

	go func() {
		err = s.listenAndServe()
		wg.Done()
	}()

	go func() {
		<-exit
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 3)
		s.Shutdown(ctx)

	}()

	wg.Wait()
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.Watchdog.Close()
	return s.server.Shutdown(ctx)
}

func (s *Server) listenAndServe() error {
	s.server = &http.Server{
		Addr:              net.JoinHostPort(s.Host, strconv.Itoa(s.Port)),
		Handler:           s.Handler,
	}

	return s.server.ListenAndServe()
}


