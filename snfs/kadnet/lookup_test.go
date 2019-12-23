package kadnet

import (
	"context"
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"net"
	"sync"
	"testing"
	"time"
)

type testLookupRpc struct {}
func (t *testLookupRpc) FindNode(c gokad.Contact, id string) (chan messages.Message, error) {
	out := make(chan messages.Message, 1)

	fnr := messages.FindNodeResponse{
		SenderID:     c.ID.String(),
		EchoRandomID: c.ID.String(),
		Payload:      []gokad.Contact{generateContact("b4945c02ddd3d4484ed7200107b46f65f5300305"), generateContact("dc03f8f281c7118225901c8655f788cd84e3f449"), generateContact("9d079f19f9edca7f8b2f5ce58624b55ffec2c4f3"), generateContact("16bcc112cd86800edfd11b0f7d2a2c476bd34f22")},
		RandomID:     c.ID.String(),
	}

	b, _ := fnr.Bytes()
	msg, _ := messages.Process(b)

	out <- msg


	return out, nil
}

type testTimeoutRpc struct {}
func (t *testTimeoutRpc) FindNode(c gokad.Contact, id string) (chan messages.Message, error) {
	out := make(chan messages.Message, 1)

	fnr := messages.FindNodeResponse{
		SenderID:     c.ID.String(),
		EchoRandomID: c.ID.String(),
		Payload:      []gokad.Contact{generateContact("b4945c02ddd3d4484ed7200107b46f65f5300305"), generateContact("dc03f8f281c7118225901c8655f788cd84e3f449"), generateContact("9d079f19f9edca7f8b2f5ce58624b55ffec2c4f3"), generateContact("16bcc112cd86800edfd11b0f7d2a2c476bd34f22")},
		RandomID:     c.ID.String(),
	}

	b, _ := fnr.Bytes()
	msg, _ := messages.Process(b)

	go func() {
		<-time.After(time.Second * 1) // trigger the push to the timeout channel
		out <- msg
	}()


	return out, nil
}

func TestNodeLookupNextRound(t *testing.T) {
	c1 := generateContact("8bc8082329609092bf86dea25cf7784cd708cc5d")
	c2 := generateContact("8f2d6ae2378dda228d3bd39c41a4b6f6f538a41a")
	c3 := generateContact("28f787e3b60f99fb29b14266c40b536d6037307e")
	contacts := map[string]gokad.Contact{
		c1.ID.String(): c1,
		c2.ID.String(): c2,
		c3.ID.String(): c3,
	}
	alphaContacts := []gokad.Contact{c1, c2, c3}
	id := "123456789"
	rpc := testLookupRpc{}
	lookup := newNodeLookup(&rpc, id, alphaContacts)
	parent := context.Background()
	ctx, _ := context.WithTimeout(parent, time.Second * 1)
	pending := make(chan pendingNode)
	timeouts := make(chan pendingNode)

	nodes := make([]pendingNode, 0)
	timedOutNodes := make([]pendingNode, 0)
	numExpectedRes := 15
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		for n := range pending {
			id := n.contact.ID.String()
			if _, ok := contacts[id]; ok {
				if !n.Queried() {
					t.Logf("Expected a seed contact's queried flag to be true %s\n", n.Contact().ID)
					continue
				}
			} else {
				if n.Queried() {
					t.Logf("Expected a non seed contac'ts queried flag to be false %s\n", n.Contact().ID)
					continue
				}
			}

			nodes = append(nodes, n)
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for n := range timeouts {
			t.Logf("Timeout node %s\n", n.contact.ID)
			timedOutNodes = append(timedOutNodes, n)
			//t.Fatalf("Did not expect a node to timeout. Node: %s\n", n.contact.ID.String())
		}

		wg.Done()
	}()

	lookup.nextRound(ctx, pending, timeouts)

	wg.Wait()
	if len(nodes) != numExpectedRes {
		t.Fatalf("Expected %d pending nodes, but got %d\n", numExpectedRes, len(nodes))
	}

	if len(timedOutNodes) > 0 {
		t.Fatalf("Expected no nodes to timeout, but got %d\n", len(timedOutNodes))
	}
}

func TestNodeLookupWithTimeout(t *testing.T) {
	c1 := generateContact("8bc8082329609092bf86dea25cf7784cd708cc5d")
	c2 := generateContact("8f2d6ae2378dda228d3bd39c41a4b6f6f538a41a")
	c3 := generateContact("28f787e3b60f99fb29b14266c40b536d6037307e")
	contacts := map[string]gokad.Contact{
		c1.ID.String(): c1,
		c2.ID.String(): c2,
		c3.ID.String(): c3,
	}
	alphaContacts := []gokad.Contact{c1, c2, c3}
	id := "123456789"
	rpc := testTimeoutRpc{}
	lookup := newNodeLookup(&rpc, id, alphaContacts)
	parent := context.Background()
	ctx, _ := context.WithTimeout(parent, time.Millisecond * 500)
	pending := make(chan pendingNode)
	timeouts := make(chan pendingNode)

	nodes := make([]pendingNode, 0)
	timedOutNodes := make([]pendingNode, 0)
	numExpectedRes := 3
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		for n := range pending {
			id := n.contact.ID.String()
			if _, ok := contacts[id]; ok {
				if !n.Queried() {
					t.Logf("Expected a seed contact's queried flag to be true %s\n", n.Contact().ID)
					continue
				}
			} else {
				if n.Queried() {
					t.Logf("Expected a non seed contac'ts queried flag to be false %s\n", n.Contact().ID)
					continue
				}
			}

			nodes = append(nodes, n)
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for n := range timeouts {
			t.Logf("Timeout node %s\n", n.contact.ID)
			timedOutNodes = append(timedOutNodes, n)
		}

		wg.Done()
	}()

	lookup.nextRound(ctx, pending, timeouts)

	wg.Wait()
	if len(nodes) != numExpectedRes {
		t.Fatalf("Expected %d pending nodes, but got %d\n", numExpectedRes, len(nodes))
	}

	if len(timedOutNodes) < 3 {
		t.Fatalf("Expected all 3 seed nodes to timeout, but only %d timed out\n", len(timedOutNodes))
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


