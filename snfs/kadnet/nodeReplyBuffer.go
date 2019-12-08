package kadnet

import (
	"errors"
	"fmt"
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
	response chan Message
}

type NodeReplyBuffer struct {
	messages    map[string]Message
	active      bool
	waitTimeout time.Duration
	// channels
	newMessage chan Message
	exit       chan bool
	subscribe  chan bufferQuery
}

func newNodeReplyBuffer() *NodeReplyBuffer {
	return &NodeReplyBuffer{
		messages:    make(map[string]Message),
		active:      false,
		waitTimeout: time.Second * 5,
		newMessage:  make(chan Message),
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

func (n *NodeReplyBuffer) Put(c Message) bool {
	if !n.IsOpen() {
		return false
	}

	n.newMessage <- c

	return true
}

func (n *NodeReplyBuffer) GetMessage(id string) (Message, error) {
	query := bufferQuery{id, make(chan Message)}
	n.subscribe <- query

	select {
	case <-time.After(n.waitTimeout):
		return Message{}, errors.New(TimeoutErr)
	case m := <-query.response:
		return m, nil
	}
}

func (n *NodeReplyBuffer) accept() {
	pending := make(map[string]chan Message)

	for {

		select {
		case m := <-n.newMessage:
			senderId := fmt.Sprintf("%x", m.SenderID)
			n.messages[senderId] = m
			c, ok := pending[senderId]
			if ok {
				c <- m
			} else {
				out := make(chan Message, 1)
				pending[senderId] = out
				out <- m

			}

		case sub := <-n.subscribe:
			c, ok := pending[sub.id]
			if ok {
				msg := <-c
				sub.response <- msg
			} else {
				out := make(chan Message, 1)
				pending[sub.id] = out
			}

		case <-n.exit:
			return
		}
	}
}
