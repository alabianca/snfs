package kadnet

import (
	"errors"
	"github.com/alabianca/gokad"
)

const WrongResponseTypeErr = "Wrong Response"

type Client struct {
	id string
	doReq chan<- *Request
}

func (c *Client) FindNode(contact gokad.Contact, lookupId string) ([]gokad.Contact, error) {
	defer c.sendPingReply(contact)
	return c.findNode(contact, lookupId)



}

func (c *Client) findNode(contact gokad.Contact, lookupId string) ([]gokad.Contact, error) {
	req := NewRequest(contact, newFindNodeRequest(c.id, "", lookupId))

	res, err := c.do(req)


	if err != nil {
		return nil, err
	}

	if res.MultiplexKey != FindNodeRes {
		return nil, errors.New(WrongResponseTypeErr)
	}

	var nodeReply FindNodeResponse
	toKademliaMessage(res, &nodeReply)

	return nodeReply.payload, nil

}

func (c *Client) do(req *Request) (*Message, error) {
	c.doReq <- req
	buf := GetNodeReplyBuffer()
	res, err := buf.GetMessage(req.Host())

	return &res, err
}

func (c *Client) sendPingReply(contact gokad.Contact) {

}


