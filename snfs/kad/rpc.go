package kad

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/kadnet"
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

type RpcManager struct {
	node *kadnet.Node
}

func NewRPCManager(dht *gokad.DHT, address string, port int) *RpcManager {
	node := kadnet.NewNode(dht, address, port)

	return &RpcManager{
		node: node,
	}
}

// RPCs start here

func (rpc *RpcManager) Bootstrap(port int, ip, idHex string) (error) {
	return rpc.node.Bootstrap(port, ip, idHex)
}


func (rpc *RpcManager) Status() []RoutingTableEntry {
	rt := make([]RoutingTableEntry, 0)

	rpc.node.Walk(func(index int, c gokad.Contact) {
		rt = append(rt, RoutingTableEntry{index, c})
	})

	return rt
}

// Manager starts here

// Service interface ID, Name, Run, Shutdown

func (rpc *RpcManager) ID() string {
	return rpc.node.ID().String()
}

func (rpc *RpcManager) Name() string {
	return ServiceName
}

func (rpc *RpcManager) Run() error {
	if err := rpc.node.Listen(nil); err != nil {
		return err
	}

	return nil
}

func (rpc *RpcManager) Shutdown() error {

	rpc.node.Shutdown()

	return nil
}
