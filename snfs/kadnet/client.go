package kadnet

import (
	"errors"
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
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
	req := NewRequest(contact, messages.NewFindNodeRequest(c.id, "", lookupId))

	res, err := c.do(req)


	if err != nil {
		return nil, err
	}

	if res.MultiplexKey != messages.FindNodeRes {
		return nil, errors.New(WrongResponseTypeErr)
	}

	var nodeReply messages.FindNodeResponse
	messages.ToKademliaMessage(res, &nodeReply)

	return nodeReply.Payload, nil

}

func (c *Client) do(req *Request) (*messages.Message, error) {
	c.doReq <- req
	buf := GetNodeReplyBuffer()
	res, err := buf.GetMessage(req.Host())

	return &res, err
}

func (c *Client) sendPingReply(contact gokad.Contact) {

}


