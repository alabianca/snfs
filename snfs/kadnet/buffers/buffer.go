package buffers

import "github.com/alabianca/snfs/snfs/kadnet/messages"

type Buffer interface {
	Close()
	BufferWriter
	BufferReader
}

type BufferReader interface {
	Read(id string, km messages.KademliaMessage) (int, error)
}

type BufferWriter interface {
	Write(msg messages.Message) (int, error)
}
