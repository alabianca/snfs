package response

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"net"
	"strconv"
)

type Response struct {
	Contact gokad.Contact
	Body buffers.BufferReader
}

func New(c gokad.Contact, reader buffers.Buffer) *Response {
	return &Response{
		Contact: c,
		Body:    reader,
	}
}

func (r *Response) Address() net.Addr {
	addr := net.JoinHostPort(r.Contact.IP.String(), strconv.Itoa(r.Contact.Port))
	updAddr, _ := net.ResolveUDPAddr("udp", addr)
	return updAddr
}

func (r *Response) Host() string {
	return r.Contact.ID.String()
}
