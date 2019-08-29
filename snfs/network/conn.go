package network

import (
	"net"
	"time"
)

type Conn struct {
	net.Conn
	IdleTimeout time.Duration
}

func (c *Conn) Write(p []byte) (int, error) {
	c.updateIdleTimeout()
	return c.Conn.Write(p)
}

func (c *Conn) Read(p []byte) (int, error) {
	c.updateIdleTimeout()
	return c.Conn.Read(p)
}

func (c *Conn) updateIdleTimeout() {
	idleDeadline := time.Now().Add(c.IdleTimeout)
	c.Conn.SetDeadline(idleDeadline)
}
