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

func nodeLookup(ownID gokad.ID, client RPC, id string, alphaCs []gokad.Contact) []gokad.Contact {
	tm := treemap.NewMap(compareDistance)
	for _, c := range alphaCs {
		node := enquiredNode{
			contact:  c,
		}
		tm.Insert(ownID.DistanceTo(c.ID), &node)
	}

	timedOutNodes := make(chan findNodeResult)
	lateReplies := losers(timedOutNodes)

	for {
		next := nextSet(tm, 3)
		for _, c := range next {
			c.queried = true
		}

		if len(next) == 0 {
			close(timedOutNodes)
			break
		}

		rc := round(client, id, next, timedOutNodes)
		for cs := range mergeLosersAndRound(lateReplies, rc) {
			if cs.err == nil {
				cs.node.answered = true
				for _, c := range cs.payload {
					tm.Insert(ownID.DistanceTo(c.ID), &enquiredNode{contact: c})
				}
			}
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
		if !node.Queried() && len(out) < size {
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
			res := <-sendFindNode(client, id, node, time.Second*3)
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