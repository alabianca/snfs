package kadmux

import (
	"net"

	"github.com/alabianca/snfs/snfs/kadnet/conn"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/request"
)

const HandlerNotFoundErr = "Handler Not Found"

type RemoteMessage interface {
	Host() string
	Address() net.Addr
}

type RpcHandler func(conn conn.KadWriter, req *request.Request)

type KadMux struct {
	reader        conn.KadReader
	handlers      map[messages.MessageType]RpcHandler
	dispatcher    *Dispatcher
	newWriterFunc func(addr net.Addr) conn.KadWriter
	// channels
	dispatchRequest chan WorkRequest
	onResponse      chan messages.Message
	onRequest       chan *request.Request
	stopReceiver    chan chan error
	stopReply       chan chan error
	stopDispatcher  chan bool
	exit            chan error
}

func NewMux() *KadMux {
	return &KadMux{
		reader:          nil,
		dispatcher:      NewDispatcher(10),
		stopDispatcher:  make(chan bool),
		dispatchRequest: make(chan WorkRequest),
		handlers:        make(map[messages.MessageType]RpcHandler),
		onRequest:       make(chan *request.Request),
		onResponse:      make(chan messages.Message),
		exit:            make(chan error),
	}
}

func (k *KadMux) Shutdown() {
	if k.stopReply != nil && k.stopReceiver != nil {
		stopReply := make(chan error)
		stopRec := make(chan error)
		k.stopReply <- stopReply
		k.stopReceiver <- stopRec

		<-stopReply
		<-stopRec
		k.stopDispatcher <- true
		k.exit <- nil
	}
}

func (k *KadMux) Start(reader conn.KadReader, nwf func(addr net.Addr) conn.KadWriter) error {
	k.reader = reader
	k.startDispatcher(10) // @todo get max workers from somewhere else
	k.newWriterFunc = nwf
	receiver := NewReceiverThread(k.onResponse, k.onRequest, k.reader)
	reply := NewReplyThread(k.onResponse, k.onRequest, k.newWriterFunc)

	k.stopReceiver = make(chan chan error)
	k.stopReply = make(chan chan error)

	go k.handleRequests()
	go receiver.Run(k.stopReceiver)
	go reply.Run(k.dispatchRequest, k.stopReply)

	return <-k.exit
}

func (k *KadMux) startDispatcher(max int) {
	for i := 0; i < max; i++ {
		w := NewWorker(i)
		k.dispatcher.Dispatch(w)
	}

	go k.dispatcher.Start()
}

func (k *KadMux) handleRequests() {
	queue := make([]*WorkRequest, 0)
	for {

		var fanout chan<- WorkRequest
		var next WorkRequest
		if len(queue) > 0 {
			fanout = k.dispatcher.QueueWork()
			next = *queue[0]
		}

		select {
		case <-k.stopDispatcher:
			k.dispatcher.Stop()
			return
		case fanout <- next:
			queue = queue[1:]
		case work := <-k.dispatchRequest:
			handler, ok := k.handlers[work.ArgRequest.MultiplexKey()]
			if !ok {
				continue
			}
			work.Handler = handler
			queue = append(queue, &work)
		}
	}
}

func (k *KadMux) HandleFunc(m messages.MessageType, handler RpcHandler) {
	k.handlers[m] = handler
}
