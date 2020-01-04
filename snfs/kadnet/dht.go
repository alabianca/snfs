package kadnet

import (
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"log"
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
	return dht.Table.ID
}

func (dht *DHT) Bootstrap(port int, ip, idHex string) (*gokad.Contact, int, error) {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()
	return dht.Table.Bootstrap(port, net.ParseIP(ip), idHex)
}

func (dht *DHT) NodeLookup(rpc RPC, id gokad.ID) {
	buf := buffers.GetNodeReplyBuffer()
	buf.Open()
	defer buf.Close()
	// contacts := make([]gokad.Contact, 0)
	alphaNodes := dht.getAlphaNodes(3, id)
	res, err := rpc.FindNode(alphaNodes[0], id.String())
	if err != nil {
		log.Printf("Error Node Lookup %s\n", err)
		return
	}
	fnr := messages.FindNodeResponse{}
	if _, err := res.Read(&fnr); err != nil {
		log.Printf("Error reading %s\n", err)
		return
	}

	log.Printf("No Error %v\n", fnr)
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
