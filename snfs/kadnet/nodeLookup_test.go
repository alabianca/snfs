package kadnet

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/repositories/treemap"
	"github.com/alabianca/snfs/snfs/kadnet/response"
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestNodeLookup_nextSet(t *testing.T) {
	id, _ := gokad.From("8f2d6ae2378dda228d3bd39c41a4b6f6f538a41a")
	c1 := generateContact("8bc8082329609092bf86dea25cf7784cd708cc5d")
	c2 := generateContact("28f787e3b60f99fb29b14266c40b536d6037307e")
	c3 := generateContact("dc03f8f281c7118225901c8655f788cd84e3f449")

	tm := treemap.NewMap(compareDistance)
	tm.Insert(id.DistanceTo(c1.ID), &enquiredNode{contact: c1})
	tm.Insert(id.DistanceTo(c2.ID), &enquiredNode{contact: c2})
	tm.Insert(id.DistanceTo(c3.ID), &enquiredNode{contact: c3})

	cases := []struct {
		IN  int
		OUT int
	}{
		{
			IN:  1,
			OUT: 1,
		},
		{
			IN:  2,
			OUT: 2,
		},
		{
			IN:  3,
			OUT: 3,
		},
		{
			IN:  4,
			OUT: 3,
		},
	}

	for _, c := range cases {
		out := nextSet(tm, c.IN)
		if len(out) != c.OUT {
			t.Fatalf("Expected %d contacts to be returned, but got %d\n", c.OUT, len(out))
		}
	}

}

func TestNodeLookup_FindNode(t *testing.T) {
	cid := "28f787e3b60f99fb29b14266c40b536d6037307a"
	contact := generateContact(cid)
	buf := buffers.NewNodeReplyBuffer()
	buf.Open()
	defer buf.Close()
	c := sendFindNode(
		newTestClient(buf, map[string]messages.FindNodeResponse{
			cid: generateFindNodeResponse(
				cid,
				"8f2d6ae2378dda228d3bd39c41a4b6f6f538a41a",
				[]gokad.Contact{generateContact("dc03f8f281c7118225901c8655f788cd84e3f449")},
			),
		}),
		"28f787e3b60f99fb29b14266c40b536d6037307e",
		&enquiredNode{contact: contact},
		time.Second*1,
	)

	out := <-c
	if out.err != nil {
		t.Fatalf("Expected error to be nil, but got %s\n", out.err)
	}

	if len(out.payload) != 1 {
		t.Fatalf("Expected 1 contact in the payload, but got %d\n", len(out.payload))
	}
}

func TestNodeLookup_FindNodeTimeout(t *testing.T) {
	cid := "28f787e3b60f99fb29b14266c40b536d6037307a"
	contact := generateContact(cid)
	buf := buffers.NewNodeReplyBuffer()
	buf.Open()
	defer buf.Close()
	c := sendFindNode(
		newTestClient(buf, map[string]messages.FindNodeResponse{
			cid: generateFindNodeResponse(
				"28f787e3b60f99fb29b14266c40b536d6037307e",
				"8f2d6ae2378dda228d3bd39c41a4b6f6f538a41e",
				[]gokad.Contact{generateContact("dc03f8f281c7118225901c8655f788cd84e3f449")},
			),
		}),
		"28f787e3b60f99fb29b14266c40b536d6037307e",
		&enquiredNode{contact: contact},
		time.Second*1,
	)

	out := <-c
	if out.err.Error() != buffers.TimeoutErr {
		t.Fatalf("Expected error to be timeout error, but got nil\n")
	}

	if len(out.payload) > 0 {
		t.Fatalf("Expected 0 contacts in the payload, but got %d\n", len(out.payload))
	}
}

func TestNodeLookup_Round(t *testing.T) {
	buf := buffers.NewNodeReplyBuffer()
	buf.Open()
	defer buf.Close()
	c1 := generateContact(gokad.GenerateRandomID().String())
	c2 := generateContact(gokad.GenerateRandomID().String())
	c3 := generateContact(gokad.GenerateRandomID().String())

	data := map[string]messages.FindNodeResponse{
		c1.ID.String(): generateFindNodeResponse(
			c1.ID.String(),
			gokad.GenerateRandomID().String(),
			[]gokad.Contact{generateContact(gokad.GenerateRandomID().String())},
		),
		c2.ID.String(): generateFindNodeResponse(
			c2.ID.String(),
			gokad.GenerateRandomID().String(),
			[]gokad.Contact{generateContact(gokad.GenerateRandomID().String())},
		),
	}

	client := newTestClient(buf, data)

	nodes := []*enquiredNode{
		{contact: c1},
		{contact: c2},
		{contact: c3},
	}

	losers := make(chan findNodeResult, 10)

	var n int
	for res := range round(client, gokad.GenerateRandomID().String(), nodes, losers) {
		n++
		if res.err != nil {
			t.Fatalf("Expected no errors from round responses, but got %s\n", res.err)
		}

		fnr, _ := data[res.node.contact.ID.String()]
		if !reflect.DeepEqual(fnr.Payload, res.payload) {
			t.Fatalf("Expected data %v\n, but got %v\n", fnr.Payload, res.payload)
		}

	}

	if n != 2 {
		t.Fatalf("Expected 2 out of 3 contacts to respond, but got %d\n", n)
	}

	var l int
	for res := range losers {
		l++
		if res.err.Error() != buffers.TimeoutErr {
			t.Fatalf("Expected timeout error but got %s\n", res.err)
		}
		break
	}

	if l != 1 {
		t.Fatalf("Expected 1 timed out node, but got %d\n", 1)
	}

}

func TestNodeLookup_MergeLosersAndRound(t *testing.T) {
	losers := make(chan findNodeResult)
	winners := make(chan findNodeResult)
	routines := runtime.NumGoroutine()

	num := make(chan int)
	go func(l, w chan findNodeResult) {
		var n int
		for range mergeLosersAndRound(l, w) {
			n++
		}
		num <- n
	}(losers, winners)

	vals := 10
	for i := 0; i < vals; i++ {
		if i % 2 == 0 {
			losers <- findNodeResult{}
		} else {
			winners <- findNodeResult{}
		}
	}

	close(winners)
	out := <- num
	if out != vals {
		t.Errorf("Expected %d values to be passed through, but got %d\n", vals, out)
	}

	nr := runtime.NumGoroutine()
	if nr != routines {
		t.Errorf("Go routines did not finish. Expected %d to run, but got %d\n", routines, nr)
	}
}

func TestNodeLookup_AllSuccess(t *testing.T) {
	lookupID := gokad.GenerateRandomID()
	c1 := generateContact("28f787e3b60f99fb29b14266c40b536d6037307e")
	c2 := generateContact("8f2d6ae2378dda228d3bd39c41a4b6f6f538a41e")
	c3 := generateContact("dc03f8f281c7118225901c8655f788cd84e3f449")
	c4 := generateContact("ac03f8f281c7118225901c8655f788cd84e3f449")
	c5 := generateContact("bc03f8f281c7118225901c8655f788cd84e3f449")
	c6 := generateContact("bc03f8f281c7118225901c8655f788cd84e3f449")
	in := map[string]gokad.Contact{
		c1.ID.String():c1,
		c2.ID.String():c2,
		c3.ID.String():c3,
		c4.ID.String():c4,
		c5.ID.String():c5,
		c6.ID.String():c6,
	}
	alphaNodes := []gokad.Contact{c1,c2,c3}
	buf := buffers.NewNodeReplyBuffer()
	buf.Open()
	defer buf.Close()

	data := map[string]messages.FindNodeResponse{
		c1.ID.String(): generateFindNodeResponse(
			c1.ID.String(),
			gokad.GenerateRandomID().String(),
			[]gokad.Contact{c4},
		),
		c2.ID.String(): generateFindNodeResponse(
			c2.ID.String(),
			gokad.GenerateRandomID().String(),
			[]gokad.Contact{c5},
		),
		c4.ID.String(): generateFindNodeResponse(
			c4.ID.String(),
			gokad.GenerateRandomID().String(),
			[]gokad.Contact{},
		),
		c5.ID.String(): generateFindNodeResponse(
			c5.ID.String(),
			gokad.GenerateRandomID().String(),
			[]gokad.Contact{c6},
		),
	}

	client := newTestClient(buf, data)

	out := nodeLookup(client, lookupID, alphaNodes)
	outm := make(map[string]gokad.Contact)
	for _, c := range out {
		outm[c.ID.String()] = c
	}

	if !reflect.DeepEqual(outm, in) {
		t.Fatalf("Expected returned nodes to be %v, but got %v\n", in, outm)
	}
}

func TestNodeLookup_NoResponse(t *testing.T) {
	lookupID := gokad.GenerateRandomID()
	c1 := generateContact("28f787e3b60f99fb29b14266c40b536d6037307e")
	c2 := generateContact("8f2d6ae2378dda228d3bd39c41a4b6f6f538a41e")
	c3 := generateContact("dc03f8f281c7118225901c8655f788cd84e3f449")

	alphaNodes := []gokad.Contact{c1,c2,c3}
	buf := buffers.NewNodeReplyBuffer()
	buf.Open()
	defer buf.Close()
	// don't add any data to the buffer to simulate no responses
	data := map[string]messages.FindNodeResponse{}

	client := newTestClient(buf, data)

	out := nodeLookup(client, lookupID, alphaNodes)
	if len(out) != 3 {
		t.Fatalf("Expected %d results, but got %d results\n", 3, len(out))
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

func generateFindNodeResponse(sender, echo string, payload []gokad.Contact) messages.FindNodeResponse {
	return messages.FindNodeResponse{
		SenderID:     sender,
		EchoRandomID: echo,
		Payload:      payload,
		RandomID:     gokad.GenerateRandomID().String(),
	}
}

type testClient struct {
	buf      buffers.Buffer
	mockData map[string]messages.FindNodeResponse
}

// creates a test RPC client and fills the buffer with the mockData
func newTestClient(buf buffers.Buffer, mockData map[string]messages.FindNodeResponse) *testClient {
	writer := buf.NewWriter()
	for _, v := range mockData {
		b, _ := v.Bytes()
		writer.Write(b)

	}

	return &testClient{
		buf:      buf,
		mockData: mockData,
	}
}

func (t *testClient) FindNode(c gokad.Contact, id string) (*response.Response, error) {
	fnr, ok := t.mockData[c.ID.String()]
	var matcher string

	if ok {
		matcher = fnr.EchoRandomID
	}

	res := response.New(c, matcher, t.buf)

	return res, nil
}
