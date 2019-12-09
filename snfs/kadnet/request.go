package kadnet

import (
	"github.com/alabianca/gokad"
	"net"
	"strconv"
)

type Request struct {
	Contact gokad.Contact
	Body KademliaMessage
}

func NewRequest(c gokad.Contact, body KademliaMessage) *Request {
	return &Request{
		Contact: c,
		Body:    body,
	}
}

func (r *Request) MultiplexKey() MessageType {
	return r.Body.MultiplexKey()
}

func (r *Request) Address() net.Addr {
	addr := net.JoinHostPort(r.Contact.IP.String(), strconv.Itoa(r.Contact.Port))
	updAddr, _ := net.ResolveUDPAddr("udp", addr)
	return updAddr
}

func (r *Request) Host() string {
	return r.Contact.ID.String()
}
