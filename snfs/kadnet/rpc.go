package kadnet

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/response"
)

const ServiceName = "RPCManager"

type RoutingTableEntry struct {
	BucketIndex int `json:"bucketIndex"`
	Contact gokad.Contact `json:"contact"`
}


type Manager interface {
	Name() string
	ID() string
	Run() error
	Shutdown() error
}

type RPC interface {
	FindNode(c gokad.Contact, id string) (*response.Response, error)
}

type RpcManager struct {
	dht    *DHT
	server *Server
}

func NewRPCManager(dht *DHT, address string, port int) *RpcManager {

	server := NewServer(dht, address, port)

	return &RpcManager{
		dht:    dht,
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
func (rpc *RpcManager) Bootstrap(port int, ip, idHex string) (error) {
	// 1. Insert gateway into k-bucket
	_, _, err := rpc.dht.Bootstrap(port, ip, idHex)
	// at capacity means we ping the head to see if it is still active. at this point contact is not inserted
	// c is the head
	if err != nil && err.Error() == gokad.ErrBucketAtCapacity {
		return err
	}
	if err != nil {
		return err
	}

	// start node lookup for own id
	ownID := rpc.dht.Table.ID
	rpc.bootstrap(ownID)

	return nil

}

// seed adds a list of contacts to the dht's routing table. Used for testing purposes only
func (rpc *RpcManager) Seed(contacts ...gokad.Contact) {
	for _, c := range contacts {
		rpc.dht.Insert(c)
	}
}

func (rpc *RpcManager) Status() []RoutingTableEntry {
	rt := make([]RoutingTableEntry, 0)

	rpc.dht.Walk(func(index int, c gokad.Contact) {
		rt = append(rt, RoutingTableEntry{index, c})
	})

	return rt
}

// @todo step 3 of the bootstrap procedure.
func (rpc *RpcManager) bootstrap(id gokad.ID) {
	client := rpc.server.NewClient()
	cs := rpc.dht.NodeLookup(client, id)
	for _, c := range cs {
		rpc.dht.Insert(c)
	}
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
