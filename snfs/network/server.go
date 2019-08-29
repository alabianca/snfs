package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type server struct {
	host        string
	port        int
	rootContext *bytes.Buffer
}

func (s *server) listen() error {
	p := strconv.Itoa(s.port)
	l, err := net.Listen("tcp", net.JoinHostPort(s.host, p))
	if err != nil {
		return err
	}

	c, err := l.Accept()
	if err != nil {
		return err
	}

	fmt.Println("got a connection")
	conn := Conn{
		Conn:        c,
		IdleTimeout: time.Second * 5,
	}
	io.Copy(&conn, s.rootContext)
	return nil
}
