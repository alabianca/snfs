package conn

import (
	"errors"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"log"
	"net"
	"sync"
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
	mtx *sync.Mutex
	conn net.PacketConn
}

func (w *writer) Write(p []byte) (int, error) {
	log.Printf("Sending: %d bytes\n", len(p))
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
	log.Printf("Written %d bytes to %s\n", written, w.addr)
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

	key, err := c.readMultiplexKey()
	if err != nil {
		return messages.Message{}, nil, err
	}

	return c.copyMessageBody(key)
}


func (c *conn) readMultiplexKey() (messages.MessageType, error) {
	firstByte := make([]byte, 1)

	_, _, err := c.conn.ReadFrom(firstByte)
	if err != nil {
		return messages.MessageType(0), err
	}

	key := messages.MessageType(firstByte[0])
	if !messages.IsValid(key) {
		log.Printf("Key %d\n", key)
		return messages.MessageType(0), errors.New(InvalidMessageTypeErr)
	}

	return key, nil
}

func (c *conn) copyMessageBody(msgType messages.MessageType) (messages.Message, net.Addr, error) {
	maxSize := messages.GetMessageSize(msgType)
	buf := make([]byte, maxSize)
	n, r, err := c.conn.ReadFrom(buf)
	if err != nil {
		return messages.Message{}, nil, err
	}

	m, err := messages.Process(buf[:n], msgType)
	if err != nil {
		return messages.Message{}, nil, err
	}

	return m, r, nil

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




