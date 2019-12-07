package kadnet

import (
	"github.com/alabianca/gokad"
	"net"
	"sync"
)

const maxMsgBuffer = 100
const ServiceName = "RPCManager"

type readResult struct {
	message Message
	remote  *net.UDPAddr
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

type Handler func(c *Conn)

type RPC interface {
	Bootstrap(port int, ip, idHex string)
}

type rpcManager struct {
	dht        *DHT
	port       int
	address    string
	conn       *net.UDPConn
	// wait groups
	mainLoops sync.WaitGroup
	// channels
	stopRead     chan bool
	stopWrite    chan bool
	doNodeLookup chan *gokad.ID
	doPing       chan *gokad.Contact
	onRequest    chan CompleteMessage
	onResponse   chan CompleteMessage

	receivedMessage chan KademliaMessage
}

func NewRPCManager(address string, port int) RPCManager {
	return &rpcManager{
		dht:        NewDHT(),
		port:       port,
		address:    address,
		mainLoops:  sync.WaitGroup{},

		stopRead:        make(chan bool),
		stopWrite:       make(chan bool),
		receivedMessage: make(chan KademliaMessage),
		doNodeLookup:    make(chan *gokad.ID),
		doPing:          make(chan *gokad.Contact),
		onRequest:       make(chan CompleteMessage),
		onResponse:      make(chan CompleteMessage),
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

	// start node lookup for own id
	//ownID := rpc.dht.Table.ID
	//nlr := newFindNodeRequest(ownID.String(), "", ownID.String())
	//rpc.onRequest <- CompleteMessage{nlr, nil}

}

// Manager starts here

// Service interface ID, Name, Run, Shutdown

func (rpc *rpcManager) ID() string {
	return rpc.dht.Table.ID.String()
}

func (rpc *rpcManager) Name() string {
	return ServiceName
}

func (rpc *rpcManager) Run() error {
	if err := rpc.Listen(); err != nil {
		return err
	}

	receiver := NewReceiverThread(rpc.onResponse, rpc.onRequest, rpc.conn, &rpc.mainLoops)
	reply := NewReplyThread(rpc.onResponse, rpc.onRequest, rpc.conn, &rpc.mainLoops)

	go receiver.Run(rpc.stopRead)
	go reply.Run(rpc.stopWrite)

	rpc.mainLoops.Wait()

	return nil
}

func (rpc *rpcManager) Shutdown() error {

	if rpc.conn != nil {
		rpc.stopWrite <- true
		rpc.stopRead <- true
		return rpc.conn.Close()
	}

	return nil
}

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
