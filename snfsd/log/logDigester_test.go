package log

import (
	"github.com/alabianca/snfs/snfsd/mock"
	"io"
	"sync"
	"testing"
	"time"
)

func TestDigester_read(t *testing.T) {
	r, w := io.Pipe()
	var ns mock.NodeService
	var digester Digester
	var wg sync.WaitGroup
	digester.NodeService = &ns
	digester.Reader = r
	digester.Writer = w

	cases := []struct{
		IN string
		OUT string
	}{
		{IN: "Hello World\n", OUT: "Hello World"},
		{IN: "Hello World Again\n", OUT: "Hello World Again"},
		{IN: "Hello World\n", OUT: "Hello World"},
		{IN: "Hello World Again\n", OUT: "Hello World Again"},
	}

	exit := make(chan struct{})

	wg.Add(2)
	go func() {
		defer wg.Done()
		var index int
		for ev := range digester.read(exit) {
			if cases[index].OUT != string(ev) {
				t.Fatalf("Expected %s, but got %s\n", cases[index], string(ev))
			}

			index++
		}
	}()

	go func() {
		defer wg.Done()
		for _, c := range cases {
			digester.Write([]byte(c.IN))
		}

		close(exit)
	}()


	wg.Wait()

}

// Ensure that buffered events are still drained even after the Digester has been closed
func TestDigester_read_withBackpressure(t *testing.T) {
	r, w := io.Pipe()
	var ns mock.NodeService
	var digester Digester
	var wg sync.WaitGroup
	digester.NodeService = &ns
	digester.Reader = r
	digester.Writer = w

	cases := []struct{
		IN string
		OUT string
	}{
		{IN: "Hello World\n", OUT: "Hello World"},
		{IN: "Hello World Again\n", OUT: "Hello World Again"},
		{IN: "Hello World\n", OUT: "Hello World"},
		{IN: "Hello World Again\n", OUT: "Hello World Again"},
	}

	exit := make(chan struct{})

	wg.Add(2)
	result := make(chan int, 1)
	go func() {
		defer wg.Done()
		var index int
		var bytesRead int
		for ev := range digester.read(exit) {
			bytesRead += len(ev)
			// some long operation that causes backpressure
			time.Sleep(time.Second * 1)
			index++
		}
		result <- bytesRead
	}()

	var bytesRead int
	go func() {
		defer wg.Done()
		for _, c := range cases {
			n, _ := digester.Write([]byte(c.IN))
			bytesRead += (n - 1) // account for the '/n' that gets lost during each write
		}

		close(exit)
	}()


	wg.Wait()
	out := <- result
	if bytesRead != out {
		t.Fatalf("Expected %d bytes to be read, but only read %d\n", bytesRead, out)
	}

}

