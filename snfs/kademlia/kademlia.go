package kademlia

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/kadnet"
	"github.com/alabianca/snfs/snfs"
	"net"
)

const tag = "Kademlia"

type node struct {
	logger *snfs.Logger
	node *kadnet.Node
}

func New(address string, port int) snfs.Network {
	n := &node{
		logger: snfs.GetLogger(),
		node: kadnet.NewNode(gokad.NewDHT(), func(n *kadnet.Node) {
			n.Host = address
			n.Port = port
		}),
	}

	return n
}

func (n *node) Bootstrap(port int, ip string) error {
	return n.node.Bootstrap(port, ip)
}

func (n *node) ID() string {
	return n.node.ID().String()
}

func (n *node) Listen() error {
	return n.node.Listen(nil)
}

func (n *node) Shutdown() {
	n.node.Shutdown()
}

func (n *node) Store(key string, ip net.IP, port int) (int, error) {
	return n.node.Store(key, ip, port)
}

func (n *node) Resolve(hash string) (net.Addr, error) {
	resolver, err := n.node.NewResolver()
	if err != nil {
		return nil, err
	}

	return resolver.Resolve(hash)
}