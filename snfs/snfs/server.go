package snfs

import (
	"fmt"
	"net"
	"strconv"
)

type server struct {
	host string
	port int
}

func (s *server) listen() error {
	p := strconv.Itoa(s.port)
	l, err := net.Listen("tcp", net.JoinHostPort(s.host, p))
	if err != nil {
		return err
	}

	_, err = l.Accept()
	if err != nil {
		return err
	}

	fmt.Println("got a connection")
	return nil
}
