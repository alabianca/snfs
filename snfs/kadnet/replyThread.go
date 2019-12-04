package kadnet

import (
	"github.com/alabianca/gokad"
	"log"
	"net"
	"sync"
)

type ReplyThread struct {
	conn *net.UDPConn
	wg         *sync.WaitGroup
	dht        *DHT
	onResponse <-chan CompleteMessage
	onRequest  <-chan CompleteMessage

	// buffers
	nodeReplyBuffer *NodeReplyBuffer
}

func NewReplyThread(res, req <-chan CompleteMessage, conn *net.UDPConn, dht *DHT, wg *sync.WaitGroup) *ReplyThread {
	wg.Add(1)
	return &ReplyThread{
		conn: conn,
		dht:             dht,
		wg:              wg,
		onRequest:       req,
		onResponse:      res,
		nodeReplyBuffer: NewNodeReplyBuffer(),
	}
}

func (r *ReplyThread) Run(exit chan bool) {
	queue := make([]CompleteMessage, 0)

	for {

		select {
		case msg := <-r.onResponse:
			r.tempStoreMsg(msg.message)
		case <-exit:
			r.wg.Done()
			return
		case msg := <-r.onRequest:
			log.Printf("Recieved Message %d\n", msg.message.MultiplexKey)
			queue = append(queue, msg)
			handler := r.nodeLookup(&msg)
			handler()
		}
	}
}

func (r *ReplyThread) tempStoreMsg(km KademliaMessage) {
	switch km.MultiplexKey() {
	case NodeLookupRes:
		r.nodeReplyBuffer.Put(km)
	}
}

func (r *ReplyThread) nodeLookup(msg *CompleteMessage) func() {
	nlr := msg.message.(*NodeLookupRequest)
	id,_ := gokad.From(nlr.payload)
	//buffer := r.nodeReplyBuffer

	return func() {
		contacts := r.dht.FindNode(*id)
		log.Printf("Found %d contacts\n", len(contacts))
	}
}
