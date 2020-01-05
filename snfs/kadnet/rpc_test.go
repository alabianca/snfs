package kadnet_test

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet"
	"net"
)

//func TestRpcManager_Bootstrap(t *testing.T) {
//	dht1 := kadnet.NewDHT()
//	dht2 := kadnet.NewDHT()
//	m1 := kadnet.NewRPCManager(dht1, "127.0.0.1", 5050)
//	m2 := kadnet.NewRPCManager(dht2, "127.0.0.1", 5051)
//	c1 := generateContact("8bc8082329609092bf86dea25cf7784cd708cc5d")
//	c2 := generateContact("28f787e3b60f99fb29b14266c40b536d6037307e")
//	m1.Seed(c1, c2)
//
//	defer shutdown(m1, m2)
//
//	go m1.Run()
//	go m2.Run()
//
//
//	contacts, _ := m2.Bootstrap(5050, "127.0.0.1", m1.ID())
//	if len(contacts) != 2 {
//		t.Fatalf("Expected %d contacts to be discovered during bootstrapping, but got %d\n", 2, len(contacts))
//	}
//
//
//
//}

func shutdown(ms ...*kadnet.RpcManager) {
	for _, m := range ms {
		m.Shutdown()
	}
}

func generateContact(id string) gokad.Contact {
	x, _ := gokad.From(id)
	return gokad.Contact{
		ID:   x,
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5050,
	}
}
