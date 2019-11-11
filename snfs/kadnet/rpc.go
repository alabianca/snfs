package kadnet

import (
	"log"
	"net"
	"time"

	"github.com/alabianca/gokad"
)

const maxMsgBuffer = 100
const ServiceName = "RPCManager"

type readResult struct {
	message Message
	err     error
}

type RPCManager interface {
	Manager
	RPC
}

type Manager interface {
	Name() string
	ID() string
	Run() error
	Listen() error
	Shutdown() error
}

type RPC interface {
	//NodeLookup()
	Bootstrap(port int, ip, idHex string)
	NodeLookup(idHex string)
}

type rpcManager struct {
	dht     *DHT
	port    int
	address string
	conn    *net.UDPConn
	// channels
	stopRead        chan bool
	stopWrite       chan bool
	doNodeLookup    chan *gokad.Contact
	doPing          chan *gokad.Contact
	receivedMessage chan Message
}

func NewRPCManager(address string, port int) RPCManager {
	return &rpcManager{
		dht:             NewDHT(),
		port:            port,
		address:         address,
		stopRead:        make(chan bool),
		stopWrite:       make(chan bool),
		receivedMessage: make(chan Message),
		doNodeLookup:    make(chan *gokad.Contact),
		doPing:          make(chan *gokad.Contact),
	}
}

// RPCs start here

// Bootstrap follows the following bootstrapping procedure
/**
	1. The gateway is inserted in the appropriate k-bucket.
	2. A node lookup for the own id is performed. Of course, the only node that will be contacted
	   initially is the gateway. Through the node lookup for the own id, the node gets to know its
	   closest neighbors.
	3. Node lookups in the range for all k-buckets with a higher index than the one of the lowest
       non-empty are performed. This fills the k-buckets of the joining node as well as communicates
       the arrival of the new node to the existing nodes. Notice that node lookups for k-buckets
       with index lower than the first non-empty would be useless, as there are no appropriate
	   contacts in the network (otherwise, the lookup for the own id would have revealed them).

@Source: Implementation of the Kademlia Hash Table by Bruno Spori Semester Thesis
https://pub.tik.ee.ethz.ch/students/2006-So/SA-2006-19.pdf
**/
func (rpc *rpcManager) Bootstrap(port int, ip, idHex string) {
	c, _, err := rpc.dht.Bootstrap(port, ip, idHex)
	// at capacity means we ping the head to see if it is still active. at this point contact is not inserted
	// c is the head
	if err != nil && err.Error() == gokad.ErrBucketAtCapacity {
		rpc.doPing <- c
		return
	}
	if err != nil {
		return
	}

	// start node lookup
	rpc.doNodeLookup <- c

}

func (rpc *rpcManager) NodeLookup(idHex string) {}

// Manager starts here

// Listen listens for udp packets
// if no error encountered, rpc.conn is set
func (rpc *rpcManager) Listen() error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: rpc.port,
		IP:   net.ParseIP(rpc.address),
	})

	if err != nil {
		return err
	}

	rpc.conn = conn
	return nil
}

func (rpc *rpcManager) ID() string {
	return rpc.dht.Table.ID.String()
}

func (rpc *rpcManager) Name() string {
	return ServiceName
}

func (rpc *rpcManager) Run() error {
	go rpc.writeLoop()
	go rpc.readLoop()

	return nil
}

func (rpc *rpcManager) Shutdown() error {
	if rpc.conn != nil {
		rpc.stopRead <- true
		rpc.stopWrite <- true
		return rpc.conn.Close()
	}

	return nil
}

func (rpc *rpcManager) readLoop() {
	if err := rpc.Listen(); err != nil {
		return
	}
	receivedMsgs := make([]Message, 0)

	var next time.Time
	var readDone chan readResult // non-nil channel means we are currently doing IO

	for {

		var readDelay time.Duration
		if now := time.Now(); next.After(now) {
			readDelay = next.Sub(now)
		}

		var startRead <-chan time.Time
		// only start reading again if currently not reading
		if readDone == nil && len(receivedMsgs) < maxMsgBuffer {
			startRead = time.After(readDelay)
		}

		var updates chan Message
		var nextMessage Message
		if len(receivedMsgs) > 0 {
			nextMessage = receivedMsgs[0]
			updates = rpc.receivedMessage
		}

		select {
		case <-rpc.stopRead:
			return

		case result := <-readDone:
			readDone = nil
			if result.err != nil {
				// try again
				next = time.Now().Add(time.Millisecond * 100)
				break
			}

			receivedMsgs = append(receivedMsgs, result.message)

		case <-startRead:

			readDone = make(chan readResult, 1)

			go func() {
				msg, err := rpc.readNextMessage()
				readDone <- readResult{msg, err}
			}()

		case updates <- nextMessage:
			receivedMsgs = receivedMsgs[1:]
		}
	}
}

func (rpc *rpcManager) writeLoop() {
	for {
		select {
		case <-rpc.stopWrite:
			return
		case c := <-rpc.doNodeLookup:
			log.Printf("Do Node Lookup %s\n", c.ID)
		case c := <-rpc.doPing:
			log.Printf("Do Ping %s\n", c.ID)
		case msg := <-rpc.receivedMessage:
			log.Printf("Received Message %d\n", msg.MultiplexKey)
		}
	}
}

func (rpc *rpcManager) readNextMessage() (Message, error) {
	msg := make([]byte, gokad.MessageSize)
	rlen, _, err := rpc.conn.ReadFromUDP(msg)
	if err != nil {
		return Message{}, err
	}

	cpy := make([]byte, rlen)
	copy(cpy, msg[:rlen])

	return process(cpy)
}
