package kadnet

import (
	"log"
	"net"
)

type WorkRequest struct {
	Handler RpcHandler
	ArgConn *net.UDPConn
	ArgMessage *Message
	ArgBuf *ReplyBuffers
}

type Worker struct {
	id int
	Work chan WorkRequest
	Workers chan chan WorkRequest
	exit chan bool
}

func NewWorker(id int) *Worker {
	return &Worker{
		id:      id,
		Work: make(chan WorkRequest),
		exit:    make(chan bool),
	}
}

func (w *Worker) Start(queue chan chan WorkRequest) {

	for {

		queue <- w.Work

		select {
		case work := <- w.Work:
			work.Handler(work.ArgConn, work.ArgBuf, work.ArgMessage)
		case <-w.exit:
			log.Printf("Exit Worker %d\n", w.id)
			return

		}
	}

}

func (w *Worker) Stop() {
	w.exit <- true
}
