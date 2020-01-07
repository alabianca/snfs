package buffers

import (
	"errors"
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"sync"
	"time"
)

const TimeoutErr = "timeout occured"
const ClosedBufferErr = "closed buffer error"

var nodeReplyBufferInstance *NodeReplyBuffer
var onceRBuf sync.Once
func GetNodeReplyBuffer() *NodeReplyBuffer {
	onceRBuf.Do(func() {
		nodeReplyBufferInstance = NewNodeReplyBuffer()
	})

	return nodeReplyBufferInstance
}

type bufferQuery struct {
	id       string
	response chan messages.Message
}

// NodeReplyBuffer stores FindNodeResponses to FindNodeRequests.
// NodeReplyBuffer implements the buffers.Buffer interface.
// When a response message is written, it is internally stored in a map
// Where the key is the id and echo random id combined.
type NodeReplyBuffer struct {
	messages    map[string]messages.Message
	active      bool
	waitTimeout time.Duration
	// channels
	newMessage chan messages.Message
	exit       chan bool
	subscribe  chan bufferQuery
}

func NewNodeReplyBuffer() *NodeReplyBuffer {
	return &NodeReplyBuffer{
		messages:    make(map[string]messages.Message),
		active:      false,
		waitTimeout: time.Second * 5,
		newMessage:  make(chan messages.Message),
		exit:        make(chan bool),
		subscribe:   make(chan bufferQuery),
	}
}

func (n *NodeReplyBuffer) Open() {
	if n.IsOpen() {
		return
	}

	n.active = true
	go n.accept()

}

func (n *NodeReplyBuffer) Close() {
	n.active = false
	n.exit <- true
}

func (n *NodeReplyBuffer) IsOpen() bool {
	return n.active
}

// Read reads the response message into msg
// Make sure that id is the combination of sender id + random id of the request that sent it.
// This ensures that we get the response to the request we sent
func (n *NodeReplyBuffer) Read(id string, msg messages.KademliaMessage, timeout time.Duration) (int, error) {
	if !n.IsOpen() {
		return 0, errors.New(ClosedBufferErr)
	}

	var exit chan<- <-chan time.Time
	if timeout != EmptyTimeout {
		exit = make(chan<- <-chan time.Time, 1)
	}

	query := bufferQuery{id, make(chan messages.Message, 1)}
	n.subscribe <- query

	select {
	case exit <- time.After(timeout):
		return 0, errors.New(TimeoutErr)
	case buf := <- query.response:
		messages.ToKademliaMessage(buf, msg)
		return len(buf), nil
	}
}

func (n *NodeReplyBuffer) Write(msg messages.Message) (int, error) {
	if !n.IsOpen() {
		return 0, errors.New(ClosedBufferErr)
	}

	n.newMessage <- msg
	return len(msg), nil
}

func (n *NodeReplyBuffer) accept() {
	pending := make(map[string]chan messages.Message)

	for {

		select {
		case m := <-n.newMessage:
			id, err := m.SenderID()
			if err != nil {
				continue
			}
			rid, err := m.EchoRandomID()
			if err != nil {
				continue
			}
			// combine sender id and echo random id to ensure reader will really get back the message we asked for
			key := id.String() + gokad.ID(rid).String()
			n.messages[key] = m
			c, ok := pending[key]
			if ok {
				c <- m
				delete(pending, key)
			} else {
				out := make(chan messages.Message, 1)
				pending[key] = out
				out <- m
			}

		case sub := <-n.subscribe:
			c, ok := pending[sub.id]
			if ok {
				msg := <-c
				sub.response <- msg
				delete(pending, sub.id)
			} else {
				pending[sub.id] = sub.response
			}

		case <-n.exit:
			return
		}
	}
}
