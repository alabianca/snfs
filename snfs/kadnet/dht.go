package kadnet

import (
	"net"

	"github.com/alabianca/gokad"
)

type DHT struct {
	Table   *gokad.DHT
	Port    int
	Address string
}

func NewDHT() *DHT {
	return &DHT{
		Table: gokad.NewDHT(),
	}
}

func (dht *DHT) GetOwnID() []byte {
	return dht.Table.ID.Bytes()
}

func (dht *DHT) Bootstrap(port int, ip, idHex string) {

	dht.Table.Bootstrap(port, net.ParseIP(ip), idHex)
}
