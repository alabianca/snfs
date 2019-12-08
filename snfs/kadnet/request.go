package kadnet

import "github.com/alabianca/gokad"

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

func (r *Request) Host() string {
	return r.Contact.ID.String()
}
