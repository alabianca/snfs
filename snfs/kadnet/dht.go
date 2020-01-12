package kadnet

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"net"
	"sync"
)

var dhtInstance *DHT
var onceDht sync.Once

// GetDHT returns a DHT singleton
func GetDHT() *DHT {
	onceDht.Do(func() {
		dhtInstance = NewDHT()
	})

	return dhtInstance
}

type DHT struct {
	Table   *gokad.DHT
	Port    int
	Address string
	mtx     sync.Mutex
}

func NewDHT() *DHT {
	return &DHT{
		Table: gokad.NewDHT(),
		mtx:   sync.Mutex{},
	}
}

func (dht *DHT) GetOwnID() []byte {
	return dht.Table.ID
}

func (dht *DHT) Bootstrap(port int, ip, idHex string) (gokad.Contact, int, error) {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()
	return dht.Table.Bootstrap(port, net.ParseIP(ip), idHex)
}

func (dht *DHT) NodeLookup(rpc RPC, id gokad.ID) []gokad.Contact {
	buf := buffers.GetNodeReplyBuffer()
	buf.Open()
	defer buf.Close()
	alphaNodes := dht.getAlphaNodes(3, id)

	return nodeLookup(rpc, id, alphaNodes)
}

func (dht *DHT) getAlphaNodes(alpha int, id gokad.ID) []gokad.Contact {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()
	return dht.Table.GetAlphaNodes(alpha, id)
}

func (dht *DHT) FindNode(id gokad.ID) []gokad.Contact {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()

	return dht.Table.FindNode(id)
}

func (dht *DHT) Insert(c gokad.Contact) (gokad.Contact, int, error) {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()
	return dht.Table.RoutingTable.Add(c)
}
