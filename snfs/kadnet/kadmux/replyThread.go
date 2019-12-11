package kadmux

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/conn"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/request"
	"net"
)

type ReplyThread struct {
	onResponse    <-chan messages.CompleteMessage
	onRequest     <-chan messages.CompleteMessage
	newWriterFunc func(addr net.Addr) conn.KadWriter
	// buffers
	nodeReplyBuffer *buffers.NodeReplyBuffer
}

func NewReplyThread(res, req <-chan messages.CompleteMessage, nwf func(addr net.Addr) conn.KadWriter) *ReplyThread {
	return &ReplyThread{
		onRequest:       req,
		onResponse:      res,
		nodeReplyBuffer: buffers.GetNodeReplyBuffer(),
		newWriterFunc:   nwf,
	}
}

func (r *ReplyThread) Run(newWork chan<- WorkRequest, exit <-chan chan error) {
	queue := make([]messages.CompleteMessage, 0)

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
			r.tempStoreMsg(msg.Message)
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

func (r *ReplyThread) tempStoreMsg(km messages.Message) {
	switch km.MultiplexKey {
	case messages.FindNodeRes:
		r.nodeReplyBuffer.Put(km)
	}
}

func (r *ReplyThread) newWorkRequest(msg messages.CompleteMessage) (WorkRequest, error) {
	km := messages.ProcessMessage(&msg.Message)
	id, err := gokad.From(messages.ToStringId(msg.Message.SenderID))
	if err != nil {
		return WorkRequest{}, err
	}
	udpAddr, err := net.ResolveUDPAddr("udp", msg.Sender.String())
	if err != nil {
		return WorkRequest{}, err
	}

	contact := gokad.Contact{
		ID:   id,
		IP:   udpAddr.IP,
		Port: udpAddr.Port,
	}

	req := request.New(contact, km)

	wReq := WorkRequest{
		ArgConn:    r.newWriterFunc(msg.Sender),
		ArgRequest: req,
	}

	return wReq, nil
}
