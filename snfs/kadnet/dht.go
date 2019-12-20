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
	return dht.Table.ID
}

func (dht *DHT) Bootstrap(port int, ip, idHex string) (*gokad.Contact, int, error) {
	dht.mtx.Lock()
	defer dht.mtx.Unlock()
	return dht.Table.Bootstrap(port, net.ParseIP(ip), idHex)
}

func (dht *DHT) NodeLookup(rpc RPC, id gokad.ID) {
	// buf := buffers.GetNodeReplyBuffer()
	// buf.Open()
	// defer buf.Close()
	// contacts := make([]gokad.Contact, 0)
	// alphaNodes := dht.getAlphaNodes(3, id)
	// strID := id.String()
	// recContacts := treemap.NewMap(compareDistances)

	// for _, c := range alphaNodes {
	// 	go func(contact gokad.Contact) {
	// 		res, err := rpc.FindNode(contact, strID)
	// 		if err == nil {
	// 			contacts = append(contacts, res...)
	// 		} else {
	// 		}
	// 	}(c)
	// }
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
