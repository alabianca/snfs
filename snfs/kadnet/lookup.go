package kadnet

import (
	"context"
	"errors"
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/repositories/treemap"
	"log"
	"sync"
)

const timeoutErr = "[lookup] timeout error"
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

// must implement repositories/treemap/PendingNode to be able to store in treemap
type pendingNode struct {
	contact  gokad.Contact
	queried  bool
	answered bool
	response chan messages.Message
}

func (q *pendingNode) Contact() gokad.Contact {
	return q.contact
}

func (q *pendingNode) Queried() bool {
	return q.queried
}

func (q *pendingNode) Answered() bool {
	return q.answered
}

func (q *pendingNode) getResponse(ctx context.Context, contacts chan<- gokad.Contact)  error {
	select {
	case <-ctx.Done():
		close(contacts)
		log.Printf("Timeout reached for %s\n", q.contact.ID)
		return errors.New(timeoutErr)
	case msg := <- q.response:
		q.answered = true
		var fnr messages.FindNodeResponse
		messages.ToKademliaMessage(&msg, &fnr)
		log.Printf("Received payload for %s -> (%d)\n", q.contact.ID, len(fnr.Payload))
		for _, c := range fnr.Payload {
			contacts <- c
		}
		close(contacts)
		return nil
	}
}

func (q *pendingNode) findNode(ctx context.Context, rpc RPC, id string, fanoutPending, fanoutTimeout chan<- pendingNode) {
	contacts := make(chan gokad.Contact)
	res, err := rpc.FindNode(q.contact, id)
	if err != nil {
		close(contacts)
		return
	}
	exit := make(chan bool)

	go func(results <-chan gokad.Contact) {
		var hasAnwer bool
		q.queried = true
		fanoutPending <- *q
		for c := range results {
			hasAnwer = true
			fanoutPending <- pendingNode{contact: c}
		}

		if !hasAnwer {
			fanoutTimeout <- *q
		}

		exit <- true

	}(contacts)

	q.response = res
	q.getResponse(ctx, contacts)
	<-exit
}

type nodesEnquired map[string]pendingNode
func (n nodesEnquired) seed(contacts ...gokad.Contact) {
	for _, c := range contacts {
		n[c.ID.String()] = pendingNode{
			contact:  c,
			queried:  false,
			answered: false,
		}
	}
}

type nodeLookup struct {
	lookupID     string
	seedContacts []gokad.Contact
	rpc          RPC
	nodes        *treemap.TreeMap
	rounds       nodesEnquired
}

func newNodeLookup(rpc RPC, id string, alphaContacts []gokad.Contact) *nodeLookup {
	l := &nodeLookup{
		seedContacts: alphaContacts,
		lookupID:     id,
		rpc:          rpc,
		nodes:        treemap.NewMap(compareDistances),
		rounds:       make(nodesEnquired),
	}

	l.rounds.seed(alphaContacts...)

	return l
}

func (n *nodeLookup) lookup() {

}

func (n *nodeLookup) nextRound(ctx context.Context, pending, timeouts chan pendingNode) {
	wg := new(sync.WaitGroup)

	for _, node := range n.rounds {
		wg.Add(1)
		go func(p pendingNode) {
			p.findNode(ctx, n.rpc, n.lookupID, pending, timeouts)
			wg.Done()
		}(node)

	}

	wg.Wait()
	close(pending)
	close(timeouts)

}
