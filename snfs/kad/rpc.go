package kad

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/kadnet"
	"net"
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
	node := kadnet.NewNode(dht, func(n *kadnet.Node) {
		n.Host = address
		n.Port = port
	})

	return &RpcManager{
		node: node,
	}
}


func (rpc *RpcManager) Bootstrap(port int, ip string) error {
	return rpc.node.Bootstrap(port, ip)
}


func (rpc *RpcManager) Status() []RoutingTableEntry {
	rt := make([]RoutingTableEntry, 0)

	rpc.node.Walk(func(index int, c gokad.Contact) {
		rt = append(rt, RoutingTableEntry{index, c})
	})

	return rt
}

func (rpc *RpcManager) Store(key string, ip net.IP, port int) (int, error) {
	return rpc.node.Store(key, ip, port)
}

func (rpc *RpcManager) Resolve(hash string) (net.Addr, error) {
	resolver, err := rpc.node.NewResolver()
	if err != nil {
		return nil, err
	}

	return resolver.Resolve(hash)
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
