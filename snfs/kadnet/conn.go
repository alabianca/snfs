package kadnet

import (
	"github.com/alabianca/gokad"
	"net"
	"sync"
)

type WriterFactory func(conn net.PacketConn) func(addr net.Addr) KadWriter

type KadConn interface {
	KadReader
	WriterFactory() func(addr net.Addr) KadWriter
	Close() error
}


type KadReader interface {
	Next() (Message, net.Addr, error)
}

type KadWriter interface {
	Write(p []byte) (int, error)
}

type writer struct {
	addr net.Addr
	mtx *sync.Mutex
	conn net.PacketConn
}

func (w *writer) Write(p []byte) (int, error) {
	return w.write(p)
}

func (w *writer) write(p []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	written := 0
	for written < len(p) {
		n, err := w.conn.WriteTo(p[written:], w.addr)
		if err != nil {
			return written,  err
		}

		written+=n
	}

	return written, nil
}

type conn struct {
	conn net.PacketConn
}

func NewConn(c net.PacketConn) KadConn {
	return &conn{
		conn: c,
	}
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) Next() (Message, net.Addr, error) {
	msg := make([]byte, gokad.MessageSize)
	rlen, raddr, err := c.conn.ReadFrom(msg)
	if err != nil {
		return Message{}, nil, err
	}

	cpy := make([]byte, rlen)
	copy(cpy, msg[:rlen])

	out, err := process(cpy)

	return out, raddr, err
}

func (c *conn) WriterFactory() func(addr net.Addr) KadWriter {
	mtx := new(sync.Mutex)
	return func(addr net.Addr) KadWriter {
		return &writer{
			addr: addr,
			mtx:  mtx,
			conn: c.conn,
		}
	}
}




