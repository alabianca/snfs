package buffers_test

import (
	"context"
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestNodeReplyBuffer_Open(t *testing.T) {
	nrb := buffers.NewNodeReplyBuffer()
	if nrb.IsOpen() {
		t.Fatalf("Expected node reply buffer to be closed initially")
	}

	nrb.Open()
	if !nrb.IsOpen() {
		t.Fatalf("Expected node reply buffer to be open after .Open")
	}

	nrb.Close()
	if nrb.IsOpen() {
		t.Fatalf("Expected node reply buffer to be closed after .Close")
	}
}

func TestNodeReplyBuffer_ReadWriteOnClosedBuffer(t *testing.T) {
	nrb := buffers.NewNodeReplyBuffer()
	_, readErr := nrb.Read("akjdflalfdj", &messages.FindNodeRequest{})
	if readErr.Error() != buffers.ClosedBufferErr {
		t.Fatalf("Expected read %s Error, but got %s\n", buffers.ClosedBufferErr, readErr.Error())
	}

	_, writeErr := nrb.Write([]byte{})
	if writeErr.Error() != buffers.ClosedBufferErr {
		t.Fatalf("Expected write %s Error, but got %s\n", buffers.ClosedBufferErr, writeErr.Error())
	}

}

func TestNodeReplyBuffer_Read(t *testing.T) {
	nrb := buffers.NewNodeReplyBuffer()
	// open the buffer for reading and writing
	nrb.Open()
	fnr := messages.FindNodeResponse{
		SenderID:     "8bc8082329609092bf86dea25cf7784cd708cc5d",
		EchoRandomID: "28f787e3b60f99fb29b14266c40b536d6037307e",
		Payload:      []gokad.Contact{generateContact("8f2d6ae2378dda228d3bd39c41a4b6f6f538a41a")},
		RandomID:     "dc03f8f281c7118225901c8655f788cd84e3f449",
	}

	msg, _ := fnr.Bytes()
	n, err := nrb.Write(msg)
	if n != len(msg) {
		t.Fatalf("Expected %d to be written, but got %d\n", len(msg), n)
	}

	if err != nil {
		t.Fatalf("Expected write error to be nil, but got %s\n", err)
	}

	fnrOut := &messages.FindNodeResponse{}
	_, rerr := nrb.Read("8bc8082329609092bf86dea25cf7784cd708cc5d" + "28f787e3b60f99fb29b14266c40b536d6037307e", fnrOut)
	if rerr != nil {
		t.Fatalf("Expected read error to be nil, but got %s\n", rerr)
	}

	if !reflect.DeepEqual(fnrOut, &fnr) {
		t.Fatalf("Expected fnrOur to equal fnr")
	}

}

func TestNodeReplyBuffer_AsyncRead(t *testing.T) {
	nrb := buffers.NewNodeReplyBuffer()
	// open the buffer for reading and writing
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	nrb.Open()
	fnr := messages.FindNodeResponse{
		SenderID:     "8bc8082329609092bf86dea25cf7784cd708cc5d",
		EchoRandomID: "28f787e3b60f99fb29b14266c40b536d6037307e",
		Payload:      []gokad.Contact{generateContact("8f2d6ae2378dda228d3bd39c41a4b6f6f538a41a")},
		RandomID:     "dc03f8f281c7118225901c8655f788cd84e3f449",
	}

	fnrOut := &messages.FindNodeResponse{}
	go func(c context.CancelFunc) {
		_, err := nrb.Read("8bc8082329609092bf86dea25cf7784cd708cc5d" + "28f787e3b60f99fb29b14266c40b536d6037307e", fnrOut)
		if err != nil {
			t.Fatalf("Expected read error to be nil, but got %s\n", err)
		}

		cancel()
	}(cancel)

	time.Sleep(time.Millisecond * 200) // let some time pass
	msg, _ := fnr.Bytes()
	nrb.Write(msg)

	<-ctx.Done()
	if !reflect.DeepEqual(fnrOut, &fnr) {
		t.Fatalf("Expected fnrOur to equal fnr")
	}
}

func TestNodeReplyBuffer_Close(t *testing.T) {
	x := runtime.NumGoroutine()
	nrb := buffers.NewNodeReplyBuffer()
	nrb.Open()
	nrb.Close()

	if nrb.IsOpen() {
		t.Fatalf("Expected buffer to be closed after .Close")
	}

	num := runtime.NumGoroutine()
	if num != x {
		t.Fatalf("Expected all go routines to be done after .Close, but %d are still running", num)
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
