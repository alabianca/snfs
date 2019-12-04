package kadnet

import (
	"github.com/alabianca/gokad"
	"net"
	"sync"
	"time"
)

type ReceiverThread struct {
	fanoutReply  chan<- CompleteMessage
	fanoutRequest chan<- CompleteMessage
	conn         *net.UDPConn
	wg           *sync.WaitGroup
}

func NewReceiverThread(res, req chan<- CompleteMessage, conn *net.UDPConn, wg *sync.WaitGroup) *ReceiverThread {
	wg.Add(1)
	return &ReceiverThread{
		fanoutReply:  res,
		fanoutRequest: req,
		wg:           wg,
		conn:         conn,
	}
}

func (r *ReceiverThread) Run(exit <-chan bool) {
	receivedMsgs := make([]*readResult, 0)
	var next time.Time
	var readDone chan readResult // non-nil channel means we are currently doing IO
	for {

		var readDelay time.Duration
		if now := time.Now(); next.After(now) {
			readDelay = next.Sub(now)
		}

		var startRead <-chan time.Time
		// only start reading again if currently not reading
		if readDone == nil && len(receivedMsgs) > maxMsgBuffer {
			startRead = time.After(readDelay)
		}

		// fanout
		var fanout chan<- CompleteMessage
		var nextMessage CompleteMessage
		if len(receivedMsgs) > 0 {
			nextMessage = CompleteMessage{processMessage(receivedMsgs[0].message), receivedMsgs[0].remote}

			if isResponse(nextMessage.message.MultiplexKey()) {
				fanout = r.fanoutReply
			} else {
				fanout = r.fanoutRequest
			}
		}


		select {
		case <-exit:
			r.wg.Done()
			return

		case result := <- readDone:
			readDone = nil
			if result.err != nil {
				// try again
				next = time.Now().Add(time.Microsecond * 100)
			} else {
				receivedMsgs = append(receivedMsgs, &result)
			}

		case <-startRead:
			readDone = make(chan readResult, 1)
			go func() {
				msg, raddr, err := r.readNextMessage()
				readDone <- readResult{msg, raddr, err}
			}()

		case fanout <- nextMessage:
			receivedMsgs = receivedMsgs[1:]
		}

	}
}

func (r *ReceiverThread) readNextMessage() (Message, *net.UDPAddr, error) {
	msg := make([]byte, gokad.MessageSize)
	rlen, raddr, err := r.conn.ReadFromUDP(msg)
	if err != nil {
		return Message{}, nil, err
	}

	cpy := make([]byte, rlen)
	copy(cpy, msg[:rlen])

	out, err := process(cpy)

	return out, raddr, err

}
