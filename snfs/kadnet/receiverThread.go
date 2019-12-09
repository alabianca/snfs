package kadnet

import (
	"log"
	"time"
)

type ReceiverThread struct {
	fanoutReply  chan<- CompleteMessage
	fanoutRequest chan<- CompleteMessage
	conn         KadReader
}

func NewReceiverThread(res, req chan<- CompleteMessage, conn KadReader) *ReceiverThread {
	return &ReceiverThread{
		fanoutReply:  res,
		fanoutRequest: req,
		conn:         conn,
	}
}

func (r *ReceiverThread) Run(exit <-chan chan error) {
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
			nextMessage = CompleteMessage{receivedMsgs[0].message, receivedMsgs[0].remote}

			if isResponse(nextMessage.message.MultiplexKey) {
				fanout = r.fanoutReply
			} else {
				fanout = r.fanoutRequest
			}
		}


		select {
		case out := <-exit:
			out <- nil
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
				msg, raddr, err := r.conn.Next()
				log.Printf("Received a message %d %s\n", msg.MultiplexKey, err)
				readDone <- readResult{msg, raddr, err}
			}()

		case fanout <- nextMessage:
			log.Printf("Fanout Received KademliaMessage %d\n", nextMessage.message.MultiplexKey)
			receivedMsgs = receivedMsgs[1:]
		}

	}
}

