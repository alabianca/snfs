package kadnet

import (
	"github.com/alabianca/gokad"
	"net"
)

type RpcHandler func(conn KadWriter, req *Request)

type ReplyThread struct {
	onResponse    <-chan CompleteMessage
	onRequest     <-chan CompleteMessage
	newWriterFunc func(addr net.Addr) KadWriter
	// buffers
	nodeReplyBuffer *NodeReplyBuffer
}

func NewReplyThread(res, req <-chan CompleteMessage, nwf func(addr net.Addr) KadWriter) *ReplyThread {
	return &ReplyThread{
		onRequest:       req,
		onResponse:      res,
		nodeReplyBuffer: GetNodeReplyBuffer(),
		newWriterFunc:   nwf,
	}
}

func (r *ReplyThread) Run(newWork chan<- WorkRequest, exit <-chan chan error) {
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
			queue = append(queue, msg)

		case fanout <- next:
			queue = queue[1:]

		}
	}
}

func (r *ReplyThread) tempStoreMsg(km Message) {
	switch km.MultiplexKey {
	case FindNodeRes:
		r.nodeReplyBuffer.Put(km)
	}
}

func (r *ReplyThread) newWorkRequest(msg CompleteMessage) (WorkRequest, error) {
	km := processMessage(&msg.message)
	id, err := gokad.From(toStringId(msg.message.SenderID))
	if err != nil {
		return WorkRequest{}, err
	}
	udpAddr, err := net.ResolveUDPAddr("udp", msg.sender.String())
	if err != nil {
		return WorkRequest{}, err
	}

	contact := gokad.Contact{
		ID:   id,
		IP:   udpAddr.IP,
		Port: udpAddr.Port,
	}

	req := NewRequest(contact, km)

	wReq := WorkRequest{
		ArgConn:    r.newWriterFunc(msg.sender),
		ArgRequest: req,
	}

	return wReq, nil
}
