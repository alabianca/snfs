package kadnet

import "net"

const HandlerNotFoundErr = "Handler Not Found"

type KadMux struct {
	reader     KadReader
	handlers map[MessageType]RpcHandler
	dispatcher *Dispatcher
	newWriterFunc func(addr net.Addr) KadWriter
	// channels
	dispatchRequest chan WorkRequest
	onResponse   chan CompleteMessage
	onRequest    chan CompleteMessage
	stopReceiver chan chan error
	stopReply    chan chan error
	stopDispatcher chan bool
	exit         chan error
}

func NewMux() *KadMux {
	return &KadMux{
		reader:       nil,
		dispatcher: NewDispatcher(10),
		stopDispatcher: make(chan bool),
		handlers:   make(map[MessageType]RpcHandler),
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
		k.stopDispatcher <- true
		k.exit <- nil
	}
}

func (k *KadMux) start(reader KadReader, nwf func(addr net.Addr) KadWriter) error {
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

		var fanout chan<-WorkRequest
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
		case work := <- k.dispatchRequest:
			handler, ok := k.handlers[work.ArgRequest.MultiplexKey()]
			if !ok {
				continue
			}

			work.Handler = handler
			queue = append(queue, &work)
		}
	}
}

func (k *KadMux) HandleFunc(m MessageType, handler RpcHandler) {
	k.handlers[m] = handler
}



