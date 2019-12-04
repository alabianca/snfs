package kadnet

type Call struct {
	ID       string
	Request  KademliaMessage
	Response chan KademliaMessage
	Done     chan bool
	Error    error
}

func newCall(message KademliaMessage) *Call {
	c := &Call{
		ID:       message.GetRandomID(),
		Request:  message,
		Done:     make(chan bool),
		Response: make(chan KademliaMessage),
	}

	return c
}
