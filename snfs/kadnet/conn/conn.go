package conn

import (
	"errors"
	"net"
	"sync"

	"github.com/alabianca/snfs/snfs/kadnet/messages"
)

const (
	InvalidMessageTypeErr = "invalid Message"
)

type WriterFactory func(conn net.PacketConn) func(addr net.Addr) KadWriter

type KadConn interface {
	KadReader
	WriterFactory() func(addr net.Addr) KadWriter
	Close() error
}

type KadReader interface {
	Next() (messages.Message, net.Addr, error)
}

type KadWriter interface {
	Write(p []byte) (int, error)
}

type writer struct {
	addr net.Addr
	mtx  *sync.Mutex
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
			return written, err
		}

		written += n
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

func (c *conn) Next() (messages.Message, net.Addr, error) {
	buf := make([]byte, messages.FindNodeResSize) // this is the largest possible message
	n, r, err := c.read(buf)
	if err != nil {
		return nil, r, err
	}

	m, err := messages.Process(buf[:n])

	return m, r, err
}

func (c *conn) read(p []byte) (int, net.Addr, error) {
	n, r, err := c.conn.ReadFrom(p)
	if err != nil {
		return n, r, err
	}

	muxKey := messages.MessageType(p[0])
	if !messages.IsValid(muxKey) {
		return n, r, errors.New(InvalidMessageTypeErr)
	}

	return n, r, err
}

// WriterFactory ensures to return a thread safe writer
// KadWriter needs to be thread safe as we are writing to it potentially from multiple
// goroutines.
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
