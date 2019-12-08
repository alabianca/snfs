package kadnet

import (
	"sync"
	"time"
)

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
	response chan KademliaMessage
}

type NodeReplyBuffer struct {
	messages    map[string]KademliaMessage
	active      bool
	waitTimeout time.Duration
	// channels
	newMessage chan KademliaMessage
	exit       chan bool
	subscribe  chan bufferQuery
}

func newNodeReplyBuffer() *NodeReplyBuffer {
	return &NodeReplyBuffer{
		messages:    make(map[string]KademliaMessage),
		active:      false,
		waitTimeout: time.Second * 5,
		newMessage:  make(chan KademliaMessage),
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

func (n *NodeReplyBuffer) Put(c KademliaMessage) bool {
	if !n.IsOpen() {
		return false
	}

	n.newMessage <- c

	return true
}

func (n *NodeReplyBuffer) getMessage(id string) KademliaMessage {
	query := bufferQuery{id, make(chan KademliaMessage)}
	n.subscribe <- query

	select {
	case <-time.After(n.waitTimeout):
		return nil
	case m := <-query.response:
		return m
	}
}

func (n *NodeReplyBuffer) accept() {
	pending := make(map[string]chan KademliaMessage)

	for {

		select {
		case m := <-n.newMessage:
			senderId := m.GetSenderID()
			n.messages[senderId] = m
			c, ok := pending[senderId]
			if ok {
				c <- m
			} else {
				out := make(chan KademliaMessage, 1)
				pending[senderId] = out
				out <- m

			}

		case sub := <-n.subscribe:
			c, ok := pending[sub.id]
			if ok {
				msg := <-c
				sub.response <- msg
			} else {
				out := make(chan KademliaMessage, 1)
				pending[sub.id] = out
			}

		case <-n.exit:
			return
		}
	}
}
