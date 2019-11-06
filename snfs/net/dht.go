package net

import "github.com/alabianca/gokad"

type DHT struct {
	Table   *gokad.DHT
	Port    int
	Address string
}

func NewDHT(rpc RPCAdapter) *DHT {
	return &DHT{
		Table:   gokad.NewDHT(),
		Address: rpc.Address(),
		Port:    rpc.Port(),
	}
}

func (dht *DHT) GetOwnID() []byte {
	return dht.Table.ID.Bytes()
}
