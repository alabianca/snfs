package client

import (
	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/buffers"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/request"
	"github.com/alabianca/snfs/snfs/kadnet/response"
)

const WrongResponseTypeErr = "Wrong Response"

type Client struct {
	ID    string
	DoReq chan<- *request.Request
}

func (c *Client) FindNode(contact gokad.Contact, lookupId string) (*response.Response, error) {
	defer c.sendPingReply(contact)
	return c.findNode(contact, lookupId)
}

func (c *Client) findNode(contact gokad.Contact, lookupId string) (*response.Response, error) {
	fnr := messages.FindNodeRequest{
		SenderID:     c.ID	,
		Payload:      lookupId,
		RandomID:     gokad.GenerateRandomID().String(),
	}

	b, err := fnr.Bytes()
	if err != nil {
		return nil, err
	}

	req := request.New(contact, b)
	res := response.New(contact, fnr.RandomID, buffers.GetNodeReplyBuffer())

	c.do(req)

	return res, nil

}

func (c *Client) do(req *request.Request) {
	c.DoReq <- req
}

func (c *Client) sendPingReply(contact gokad.Contact) {

}
