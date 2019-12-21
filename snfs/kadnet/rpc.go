package kadnet

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
)

const ServiceName = "RPCManager"

type Manager interface {
	Name() string
	ID() string
	Run() error
	Shutdown() error
}

type RPC interface {
	FindNode(c gokad.Contact, id string) (chan messages.Message, error)
}

type RpcManager struct {
	dht    *DHT
	server *Server
}

func NewRPCManager(address string, port int) *RpcManager {

	server := NewServer(GetDHT(), address, port)

	return &RpcManager{
		dht:    GetDHT(),
		server: server,
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
func (rpc *RpcManager) Bootstrap(port int, ip, idHex string) {
	// 1. Insert gateway into k-bucket
	_, _, err := rpc.dht.Bootstrap(port, ip, idHex)
	// at capacity means we ping the head to see if it is still active. at this point contact is not inserted
	// c is the head
	if err != nil && err.Error() == gokad.ErrBucketAtCapacity {
		return
	}
	if err != nil {
		return
	}

	// start node lookup for own id
	ownID := rpc.dht.Table.ID
	rpc.bootstrap(ownID)
	//nlr := newFindNodeRequest(ownID.String(), "", ownID.String())
	//rpc.onRequest <- CompleteMessage{nlr, nil}

}

func (rpc *RpcManager) bootstrap(id gokad.ID) {
	client := rpc.server.NewClient()
	rpc.dht.NodeLookup(client, id)
}

// Manager starts here

// Service interface ID, Name, Run, Shutdown

func (rpc *RpcManager) ID() string {
	return rpc.dht.Table.ID.String()
}

func (rpc *RpcManager) Name() string {
	return ServiceName
}

func (rpc *RpcManager) Run() error {

	if err := rpc.server.Listen(); err != nil {
		return err
	}

	return nil
}

func (rpc *RpcManager) Shutdown() error {

	rpc.server.Shutdown()

	return nil
}
