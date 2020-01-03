package request

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"net"
	"strconv"
)

type Request struct {
	Contact gokad.Contact
	Body    messages.Message
}

func New(c gokad.Contact, body messages.Message) *Request {
	return &Request{
		Contact: c,
		Body:    body,
	}
}

func (r *Request) MultiplexKey() messages.MessageType {
	key, _ := r.Body.MultiplexKey()
	return key
}

func (r *Request) Address() net.Addr {
	addr := net.JoinHostPort(r.Contact.IP.String(), strconv.Itoa(r.Contact.Port))
	updAddr, _ := net.ResolveUDPAddr("udp", addr)
	return updAddr
}

func (r *Request) Host() string {
	return r.Contact.ID.String()
}
