package kadnet

import (
	"net"
	"sync"
)

type KadMux struct {
	conn     *net.UDPConn
	handlers map[MessageType]RpcHandler
	wg       sync.WaitGroup
	// channels
	onResponse   chan CompleteMessage
	onRequest    chan CompleteMessage
	stopReceiver chan chan error
	stopReply    chan chan error
	exit         chan error
}

func NewMux() *KadMux {
	return &KadMux{
		conn:       nil,
		handlers:   make(map[MessageType]RpcHandler),
		wg:         sync.WaitGroup{},
		onRequest:  make(chan CompleteMessage),
		onResponse: make(chan CompleteMessage),
		exit:       make(chan error),
	}
}

func (k *KadMux) shutdown() {
	if k.stopReply != nil && k.stopReceiver != nil {
		stopReply := make(chan error)
		stopRec := make(chan error)
		k.stopReply <- stopReply
		k.stopReceiver <- stopRec

		<-stopReply
		<-stopRec
		k.exit <- nil
	}
}

func (k *KadMux) start(conn *net.UDPConn) error {
	k.conn = conn // @todo create custom conn here
	receiver := NewReceiverThread(k.onResponse, k.onRequest, k.conn)
	reply := NewReplyThread(k.onResponse, k.onRequest, k.conn)

	k.stopReceiver = make(chan chan error)
	k.stopReply = make(chan chan error)
	go receiver.Run(k.stopReceiver)
	go reply.Run(k.stopReply)

	return <-k.exit
}

func (k *KadMux) HandleFunc(m MessageType, handler RpcHandler) {
	k.handlers[m] = handler
}
