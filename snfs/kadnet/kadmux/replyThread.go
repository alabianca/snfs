package kadmux

import (
	"net"

	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/conn"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/request"
)

type ReplyThread struct {
	onResponse    <-chan messages.Message
	onRequest     <-chan *request.Request
	newWriterFunc func(addr net.Addr) conn.KadWriter
	// buffers
	nodeReplyBuffer *buffers.NodeReplyBuffer
}

func NewReplyThread(res chan messages.Message, req <-chan *request.Request, nwf func(addr net.Addr) conn.KadWriter) *ReplyThread {
	return &ReplyThread{
		onRequest:       req,
		onResponse:      res,
		nodeReplyBuffer: buffers.GetNodeReplyBuffer(),
		newWriterFunc:   nwf,
	}
}

func (r *ReplyThread) Run(newWork chan<- WorkRequest, exit <-chan chan error) {
	queue := make([]*request.Request, 0)

	for {

		var next WorkRequest
		var fanout chan<- WorkRequest
		if len(queue) > 0 {
			next = r.newWorkRequest(queue[0])
			fanout = newWork
		}

		select {
		case msg := <-r.onResponse:
			r.tempStoreMsg(msg)
		case out := <-exit:
			out <- nil
			return
		case req := <-r.onRequest:
			queue = append(queue, req)

		case fanout <- next:
			queue = queue[1:]

		}
	}
}

func (r *ReplyThread) tempStoreMsg(km messages.Message) {
	key, _ := km.MultiplexKey()
	buf := r.getBuffer(key)

	writer := buf.NewWriter()
	writer.Write(km)
}

func (r *ReplyThread) getBuffer(key messages.MessageType) buffers.Buffer {
	var buf buffers.Buffer
	switch key {
	case messages.FindNodeRes:
		buf = r.nodeReplyBuffer
	}

	return buf
}

func (r *ReplyThread) newWorkRequest(req *request.Request) WorkRequest {
	wReq := WorkRequest{
		ArgConn:    r.newWriterFunc(req.Address()),
		ArgRequest: req,
	}

	return wReq
}
