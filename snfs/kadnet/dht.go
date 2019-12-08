package kadnet

import (
	"net"
	"sync"

	"github.com/alabianca/gokad"
)

var dhtInstance *DHT
var onceDht sync.Once

// GetDHT returns a DHT singleton
func GetDHT() *DHT {
	onceDht.Do(func() {
		dhtInstance = newDHT()
	})

	return dhtInstance
}

type DHT struct {
	Table   *gokad.DHT
	Port    int
	Address string
	mtx     sync.Mutex
}

func newDHT() *DHT {
	return &DHT{
		Table: gokad.NewDHT(),
		mtx:   sync.Mutex{},
	}
}

func (dht *DHT) GetOwnID() []byte {
	return dht.Table.ID.Bytes()
}

func (dht *DHT) Bootstrap(port int, ip, idHex string) (*gokad.Contact, int, error) {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()
	return dht.Table.Bootstrap(port, net.ParseIP(ip), idHex)
}


func (dht *DHT) FindNode(id gokad.ID) []gokad.Contact {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()

	return dht.Table.FindNode(id)
}
