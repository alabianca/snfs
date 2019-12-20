package client

import (
	"errors"

	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/request"
)

const WrongResponseTypeErr = "Wrong Response"

type Client struct {
	ID    string
	DoReq chan<- *request.Request
}

func (c *Client) FindNode(contact gokad.Contact, lookupId string) (messages.FindNodeResponse, error) {
	defer c.sendPingReply(contact)
	return c.findNode(contact, lookupId)
}

func (c *Client) findNode(contact gokad.Contact, lookupId string) (messages.FindNodeResponse, error) {
	req := request.New(contact, messages.NewFindNodeRequest(c.ID, "", lookupId))

	res, err := c.do(req)

	if err != nil {
		return messages.FindNodeResponse{}, err
	}

	if res.MultiplexKey != messages.FindNodeRes {
		return messages.FindNodeResponse{}, errors.New(WrongResponseTypeErr)
	}

	var nodeReply messages.FindNodeResponse
	messages.ToKademliaMessage(res, &nodeReply)

	return nodeReply, nil

}

func (c *Client) do(req *request.Request) (*messages.Message, error) {
	c.DoReq <- req
	buf := buffers.GetNodeReplyBuffer()
	res, err := buf.GetMessage(req.Host())

	return &res, err
}

func (c *Client) sendPingReply(contact gokad.Contact) {

}
