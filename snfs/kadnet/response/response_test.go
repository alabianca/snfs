package response

import (
	"errors"
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"net"
	"testing"
	"time"
)

func TestResponse_ReadWithTimeout(t *testing.T) {
	res := New(
		generateContact("8bc8082329609092bf86dea25cf7784cd708cc5d"),
		"12345",
		&testBuffer{time.Millisecond * 100},
	)

	// don't allow enough time to respond
	res.ReadTimeout(time.Millisecond * 50)

	var fnr messages.FindNodeResponse

	n, err := res.Read(&fnr)
	if n > 0 {
		t.Fatalf("Expected n to be 0, but got %d\n", n)
	}

	if err == nil {
		t.Fatalf("Expected error to be timeout, but got nil")
	}

	if res.readTimeout != time.Duration(0) {
		t.Fatalf("Expected timeout to be reset after every read, but got %s\n", res.readTimeout)
	}
}

func TestResponse_ReadWithoutTimeout(t *testing.T) {
	res := New(
		generateContact("8bc8082329609092bf86dea25cf7784cd708cc5d"),
		"12345",
		&testBuffer{time.Millisecond * 100},
	)


	var fnr messages.FindNodeResponse

	n, err := res.Read(&fnr)
	if err != nil {
		t.Fatalf("Expected error to be nil, but got %s\n", err)
	}
	if n == 0 {
		t.Fatalf("Expected n to be larger than 0 after successfull read, but got %d\n", n)
	}

	if res.readTimeout != time.Duration(0) {
		t.Fatalf("Expected timeout to be reset after every read, but got %s\n", res.readTimeout)
	}
}

type testBuffer struct {
	mockTimeout time.Duration
}

func (t *testBuffer) Close() {}
func (t *testBuffer) Read(id string, km messages.KademliaMessage, timeout time.Duration) (int, error) {
	out := make(chan messages.Message)
	var exit chan<- <-chan time.Time
	if timeout != buffers.EmptyTimeout {
		exit = make(chan<- <-chan time.Time, 1)
	}

	go func() {
		// mock the reader waiting to read the response
		<-time.After(t.mockTimeout)
		// create a fake findNodeResponse
		fnr := messages.FindNodeResponse{
			SenderID:     "8bc8082329609092bf86dea25cf7784cd708cc5d",
			EchoRandomID: "8f2d6ae2378dda228d3bd39c41a4b6f6f538a41a",
			Payload:      []gokad.Contact{},
			RandomID:     "28f787e3b60f99fb29b14266c40b536d6037307e",
		}

		data, _ := fnr.Bytes()
		out <- data
	}()

	var n int
	var err error
	select {
	case msg := <-out:
		messages.ToKademliaMessage(msg, km)
		n = len(msg)
	case exit <- time.After(timeout):
		n = 0
		err = errors.New("timeout")

	}

	return n, err
}

func (t *testBuffer) Write(msg messages.Message) (int, error) {
	return 0, nil
}

func generateContact(id string) gokad.Contact {
	x, _ := gokad.From(id)
	return gokad.Contact{
		ID:   x,
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5050,
	}
}
