package net

type RPCAdapter interface {
	Port() int
	Address() string
}

func NewUDPRPCAdapter(port int, address string) RPCAdapter {
	return &udpRpc{
		port:    port,
		address: address,
	}
}

type udpRpc struct {
	port    int
	address string
}

func (u *udpRpc) Port() int {
	return u.port
}

func (u *udpRpc) Address() string {
	return u.address
}
