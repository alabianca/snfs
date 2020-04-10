package snfs

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)


type Server struct {
	Handler http.Handler
	Host string
	Port int
	server *http.Server
}

func (s *Server) Run(exit chan struct{}) error {
	var wg sync.WaitGroup
	wg.Add(2)
	var err error
	go func() {
		err = s.listenAndServe()
		wg.Done()
	}()

	go func() {
		<-exit
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 2)
		s.Shutdown(ctx)
		wg.Done()
	}()

	wg.Wait()
	return err
}

func (s *Server) listenAndServe() error {
	s.server = &http.Server{
		Addr:              net.JoinHostPort(s.Host, strconv.Itoa(s.Port)),
		Handler:           s.Handler,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
