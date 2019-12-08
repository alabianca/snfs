package kadnet

import (
	"github.com/alabianca/gokad"
	"net"
	"sync"
)

type Conn struct {
	conn net.PacketConn
	mtx  sync.Mutex
}

func NewConn(c net.PacketConn) *Conn {
	return &Conn{
		conn: c,
		mtx:  sync.Mutex{},
	}
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Next() (Message, net.Addr, error) {
	msg := make([]byte, gokad.MessageSize)
	rlen, raddr, err := c.ReadFromUDP(msg)
	if err != nil {
		return Message{}, nil, err
	}

	cpy := make([]byte, rlen)
	copy(cpy, msg[:rlen])

	out, err := process(cpy)

	return out, raddr, err
}

func (c *Conn) ReadFromUDP(p []byte) (int, net.Addr, error) {
	return c.conn.ReadFrom(p)
}

func (c *Conn) WriteAll(p []byte, addr net.Addr) (error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	written := 0
	for written < len(p) {
		n, err := c.conn.WriteTo(p[written:], addr)
		if err != nil {
			return err
		}

		written+=n
	}

	return nil
}


