package kadnet

type Call struct {
	ID       string
	Request  KademliaMessage
	Response KademliaMessage
	Done     chan bool
	Error    error
}

func newCall(request KademliaMessage) *Call {
	c := &Call{
		ID:      request.GetRandomID(),
		Request: request,
		Done:    make(chan bool),
	}

	return c
}
