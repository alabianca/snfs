package buffers

import (
	"errors"
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

func (n *NodeReplyBuffer) Read(id string, msg messages.KademliaMessage) (int, error) {
	if !n.IsOpen() {
		return 0, errors.New(ClosedBufferErr)
	}

	query := bufferQuery{id, make(chan messages.Message, 1)}
	n.subscribe <- query

	buf := <- query.response
	messages.ToKademliaMessage(buf, msg)

	return len(buf), nil
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
			senderId := id.String()
			n.messages[senderId] = m
			c, ok := pending[senderId]
			if ok {
				c <- m
				delete(pending, senderId)
			} else {
				out := make(chan messages.Message, 1)
				pending[senderId] = out
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
