package buffers

import (
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"time"
)

const EmptyTimeout = time.Duration(0)

type Buffer interface {
	Close()
	BufferWriter
	BufferReader
}

type BufferReader interface {
	Read(id string, km messages.KademliaMessage, timeout time.Duration) (int, error)
}

type BufferWriter interface {
	Write(msg messages.Message) (int, error)
}
