package kadnet

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/repositories/treemap"
	"github.com/alabianca/snfs/snfs/kadnet/response"
	"sync"
	"time"
)

const K = 20

var compareDistance = func(d1 gokad.Distance, d2 gokad.Distance) int {
	for i := 0; i < gokad.SIZE; i++ {
		if d1[i] > d2[i] {
			return -1
		}
		if d1[i] < d2[i] {
			return 1
		}
	}

	return 1
}

type enquiredNode struct {
	contact  gokad.Contact
	queried  bool
	answered bool
}

func (e *enquiredNode) Contact() gokad.Contact {
	return e.contact
}
func (e *enquiredNode) Queried() bool {
	return e.queried
}
func (e *enquiredNode) Answered() bool {
	return e.answered
}

// The node lookup is central to Kademlia.
// It works as follows:
// 1. We insert the alpha contacts in a list. This list is sorted according to
//    distance of the id we look up and the contact's id.
// 2. We then split up the node lookup into rounds. At the start of each round
//    we pull out CONCURRENCY (3) contacts ouf of the list that have not yet been queried
// 3. We then send FIND_NODE_RPC's to each of these contacts. Giving them a specified amount of time to respond
//    If they respond on time we insert each contact in the response into the list mentioned above.
//    If they don't respond on time we put the contact in a losers list and wait for a response without a timeout. If they end up responding. Great. We add it to the list
// 4. After each round is done, we again take CONCURRENCY (3) contacts and start the process again.
// 5. If a round did not reveal a new contact however, we set CONCURRENCY to K and send each of them a FIND_NODE_RPC in hopes to discover more nodes.
/*

   -----------------------|
   |	                  |
   |      (2) ------ (3)  |
   |	 /             \  |
   |	/				(5)<--- |
   |   /               /        |
   |> (1) ---(2)------ (3)      |
      \                         |
	   \	                    |
        \                       |
		  (2) ------ (3) ----- (4)


   (1): nodeLookup
   (2): sendFindNode
   (3): Wait for response with timeout
   (4): timed out nodes. Wait for response indefinetely
   (5): Gather responses from current round of timed out nodes

*/
func nodeLookup(client RPC, id gokad.ID, alphaCs []gokad.Contact) []gokad.Contact {
	tm := treemap.NewMap(compareDistance)
	for _, c := range alphaCs {
		node := enquiredNode{
			contact:  c,
		}
		tm.Insert(id.DistanceTo(c.ID), &node)
	}

	timedOutNodes := make(chan findNodeResult)
	lateReplies := losers(timedOutNodes)
	concurrency := 3
	for {
		next := nextSet(tm, concurrency)

		if len(next) == 0 {
			close(timedOutNodes)
			break
		}

		for _, c := range next {
			c.queried = true
		}

		rc := round(client, id.String(), next, timedOutNodes)
		var atLeastOneNewNode bool
		for cs := range mergeLosersAndRound(lateReplies, rc) {
			if cs.err == nil {
				cs.node.answered = true
				for _, c := range cs.payload {
					distance := id.DistanceTo(c.ID)
					if _, ok := tm.Get(distance); !ok {
						atLeastOneNewNode = true
						tm.Insert(distance, &enquiredNode{contact: c})
					}

				}
			}
		}

		// if a round did not reveal at least one new node we take all K closest
		// nodes not already queried and send them FIND_NODE_RPC's
		if !atLeastOneNewNode {
			concurrency = K
		}
	}

	out := make([]gokad.Contact, K) // @todo get this value from somewhere else instead of hardcoding it
	var index int
	tm.Traverse(func(k gokad.Distance, v treemap.PendingNode) bool {
		out[index] = v.Contact()
		index++
		if index >= K {
			return false
		}
		return true
	})

	return out[:index]
}

func nextSet(tm *treemap.TreeMap, size int) []*enquiredNode {
	out := make([]*enquiredNode, 0)

	tm.Traverse(func(k gokad.Distance, node treemap.PendingNode) bool {
		if !node.Queried() && (len(out) < size || size < 0) {
			p, _ := node.(*enquiredNode)
			out = append(out, p)
		}

		if len(out) >= size {
			return false
		}

		return true
	})

	return out
}

type findNodeResult struct {
	node    *enquiredNode
	payload []gokad.Contact
	response *response.Response
	err     error
}

func mergeLosersAndRound(losers, round <-chan findNodeResult) chan findNodeResult {
	out := make(chan findNodeResult)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case res := <- losers:
				out <- res
			case res, ok := <- round:
				if !ok {
					return
				}

				out <- res
			}
		}
	}()


	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func losers(timedOutNodes <-chan findNodeResult) chan findNodeResult {
	out := make(chan findNodeResult)
	go func() {
		for res := range timedOutNodes {
			go readWithoutTimeout(res, out)
		}
		close(out)
	}()

	return out
}

func round(client RPC, id string, nodes []*enquiredNode, losers chan findNodeResult) chan findNodeResult {
	out := make(chan findNodeResult)
	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for _, n := range nodes {
		go func(node *enquiredNode) {
			defer wg.Done()
			res := <-sendFindNode(client, id, node, time.Millisecond*500) // @todo get this timeout from somehwere
			if res.err != nil && res.err.Error() == buffers.TimeoutErr {
				losers <- res
				return
			}
			out <- res
		}(n)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out

}

func sendFindNode(client RPC, id string, node *enquiredNode, timeout time.Duration) chan findNodeResult {
	out := make(chan findNodeResult)
	go func() {
		defer close(out)
		res, err := client.FindNode(node.contact, id)
		if err != nil {
			out <- findNodeResult{node, nil, res, err}
			return
		}

		var fnr messages.FindNodeResponse
		res.ReadTimeout(timeout)
		if _, err := res.Read(&fnr); err != nil {
			out <- findNodeResult{node, nil, res, err}
			return
		}

		out <- findNodeResult{node, fnr.Payload, res, err}

	}()

	return out
}

func readWithoutTimeout(timedOutNode findNodeResult, out chan findNodeResult) {
	var fnr messages.FindNodeResponse
	if _, err := timedOutNode.response.Read(&fnr); err == nil {
		timedOutNode.err = nil
		timedOutNode.payload = fnr.Payload
		out <- timedOutNode
	}
}