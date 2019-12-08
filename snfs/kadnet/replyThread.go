package kadnet

import (
	"log"
)

type RpcHandler func(conn *Conn, req *Message)

type ReplyThread struct {
	conn       *Conn
	onResponse <-chan CompleteMessage
	onRequest  <-chan CompleteMessage

	// buffers
	nodeReplyBuffer *NodeReplyBuffer
}

func NewReplyThread(res, req <-chan CompleteMessage, conn *Conn) *ReplyThread {
	return &ReplyThread{
		conn:            conn,
		onRequest:       req,
		onResponse:      res,
		nodeReplyBuffer: GetNodeReplyBuffer(),
	}
}


func (r *ReplyThread) Run(newWork chan<-WorkRequest, exit <-chan chan error) {
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

			fanout = newWork
		}

		select {
		case msg := <-r.onResponse:
			r.tempStoreMsg(msg.message)
		case out := <-exit:
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
		r.nodeReplyBuffer.Put(km)
	}
}

func (r *ReplyThread) newWorkRequest(msg CompleteMessage) (WorkRequest, error) {
	req :=  WorkRequest{
		ArgConn:    r.conn,
		ArgMessage: &msg.message,
	}

	return req, nil
}
