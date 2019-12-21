package client

import (
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

func (c *Client) FindNode(contact gokad.Contact, lookupId string) (chan messages.Message, error) {
	defer c.sendPingReply(contact)
	return c.findNode(contact, lookupId)
}

func (c *Client) findNode(contact gokad.Contact, lookupId string) (chan messages.Message, error) {
	req := request.New(contact, messages.NewFindNodeRequest(c.ID, "", lookupId))

	return c.do(req)

}

func (c *Client) do(req *request.Request) (chan messages.Message, error) {
	c.DoReq <- req
	buf := buffers.GetNodeReplyBuffer()
	return  buf.GetMessage(req.Host())
}

func (c *Client) sendPingReply(contact gokad.Contact) {

}
