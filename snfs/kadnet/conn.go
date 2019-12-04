package kadnet

import (
	"net"
	"sync"
)

// @todo: add timeout mechanism
type Conn struct {
	conn *net.UDPConn
	mtx  sync.Mutex
}

func NewConn(c *net.UDPConn) *Conn {
	return &Conn{
		conn: c,
		mtx:  sync.Mutex{},
	}
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) SendFull(p []byte, r *net.UDPAddr) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	var offset int
	var err error

	for offset < len(p) {
		offset, err = c.conn.WriteTo(p[offset:len(p)], r)
		if err != nil {
			return err
		}

		offset += offset

	}

	return nil
}
