package kadnet

import (
	"errors"
	"log"
	"net"
)

type RpcHandler func(conn *net.UDPConn, buf *ReplyBuffers, req *Message)

type ReplyThread struct {
	conn       *net.UDPConn
	onResponse <-chan CompleteMessage
	onRequest  <-chan CompleteMessage
	handlers   map[MessageType]RpcHandler
	// thread pool
	dispatcher *Dispatcher

	// buffers
	nodeReplyBuffer *NodeReplyBuffer
}

func NewReplyThread(res, req <-chan CompleteMessage, conn *net.UDPConn, h map[MessageType]RpcHandler) *ReplyThread {
	return &ReplyThread{
		conn:            conn,
		onRequest:       req,
		onResponse:      res,
		handlers:        h,
		dispatcher:      NewDispatcher(),
		nodeReplyBuffer: NewNodeReplyBuffer(),
	}
}

func (r *ReplyThread) StartDispatcher(max int) {
	for i := 0; i < max; i++ {
		w := NewWorker(i)
		r.dispatcher.Dispatch(w)
	}

	go r.dispatcher.Start()
}

func (r *ReplyThread) Run(exit <-chan chan error) {
	queue := make([]CompleteMessage, 0)

	for {

		var next WorkRequest
		var fanout chan<- WorkRequest
		var err error
		if len(queue) > 0 {
			next, err = r.newWorkRequest(queue[0])
			if err != nil {
				queue = queue[1:]
				continue
			}

			fanout = r.dispatcher.QueueWork()
		}

		select {
		case msg := <-r.onResponse:
			r.tempStoreMsg(msg.message)
		case out := <-exit:
			r.dispatcher.Stop()
			out <- nil
			return
		case msg := <-r.onRequest:
			log.Printf("Recieved Message %d\n", msg.message.MultiplexKey)
			queue = append(queue, msg)

		case fanout <- next:
			queue = queue[1:]

		}
	}
}

func (r *ReplyThread) tempStoreMsg(km Message) {
	switch km.MultiplexKey {
	case FindNodeRes:
		log.Printf("Store Lookup Response %s\n", km.SenderID)
		r.nodeReplyBuffer.Put(processMessage(km))
	}
}

func (r *ReplyThread) newWorkRequest(msg CompleteMessage) (WorkRequest, error) {
	handler, ok := r.handlers[msg.message.MultiplexKey]
	if !ok {
		return WorkRequest{}, errors.New(HandlerNotFoundErr)
	}
	buf := &ReplyBuffers{nodeReplyBuffer: r.nodeReplyBuffer}
	req :=  WorkRequest{
		Handler:    handler,
		ArgConn:    r.conn,
		ArgMessage: &msg.message,
		ArgBuf:     buf,
	}

	return req, nil
}
