package buffers

import (
	"errors"
	"fmt"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"sync"
	"time"
)

const TimeoutErr = "Timeout Occured"

var nodeReplyBufferInstance *NodeReplyBuffer
var onceRBuf sync.Once
func GetNodeReplyBuffer() *NodeReplyBuffer {
	onceRBuf.Do(func() {
		nodeReplyBufferInstance = newNodeReplyBuffer()
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

func newNodeReplyBuffer() *NodeReplyBuffer {
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

func (n *NodeReplyBuffer) Put(c messages.Message) bool {
	if !n.IsOpen() {
		return false
	}

	n.newMessage <- c

	return true
}

func (n *NodeReplyBuffer) GetMessage(id string) (messages.Message, error) {
	query := bufferQuery{id, make(chan messages.Message)}
	n.subscribe <- query

	select {
	case <-time.After(n.waitTimeout):
		return messages.Message{}, errors.New(TimeoutErr)
	case m := <-query.response:
		return m, nil
	}
}

func (n *NodeReplyBuffer) accept() {
	pending := make(map[string]chan messages.Message)

	for {

		select {
		case m := <-n.newMessage:
			senderId := fmt.Sprintf("%x", m.SenderID)
			n.messages[senderId] = m
			c, ok := pending[senderId]
			if ok {
				c <- m
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
			} else {
				out := make(chan messages.Message, 1)
				pending[sub.id] = out
			}

		case <-n.exit:
			return
		}
	}
}
