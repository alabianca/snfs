package kadnet

import (
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
			log.Printf("Recieved KademliaMessage %d\n", msg.message.MultiplexKey)
			queue = append(queue, msg)

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

//func (r *ReplyThread) nodeLookup(msg *CompleteMessage) func() {
//	nlr := msg.message.(*FindNodeRequest)
//	id,_ := gokad.From(nlr.payload)
//	//buffer := r.nodeReplyBuffer
//
//	return func() {
//		contacts := r.dht.FindNode(*id)
//		log.Printf("Found %d contacts\n", len(contacts))
//	}
//}
