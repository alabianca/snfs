package kadnet

import (
	"github.com/alabianca/gokad"
)

// if d1 is larger than d2 return 1
// if d2 is larger return -1
// if they are the same return 0
var compareDistances = func(d1, d2 gokad.Distance) int {
	for i := 0; i < gokad.MaxCapacity; i++ {
		if d1[i] > d2[i] {
			return 1
		}
		if d1[i] < d2[i] {
			return -1
		}
	}

	return 1
}

type queriedNode struct {
	contact  gokad.Contact
	queried  bool
	answered bool
}

type nodesEnquired map[string]*queriedNode

type nodeLookup struct {
	seedID       string
	seedContacts []gokad.Contact
	rpc          RPC
}

func newNodeLookup(rpc RPC, seedID string, alphaContacts []gokad.Contact) *nodeLookup {
	l := &nodeLookup{
		seedContacts: alphaContacts,
		seedID:       seedID,
		rpc:          rpc,
	}

	return l
}

func (n *nodeLookup) lookup() {

}
