package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type NodeService struct {
	api *RestAPI
}

func NewNodeService() *NodeService {
	return &NodeService{api:NewRestAPI(getBaseURL())}
}

func (n *NodeService) Create(cport, dport int, name string) (NodeConfiguration, error) {
	c := NodeConfiguration{
		Name:      name,
		Cport:     cport,
		Dport:     dport,
	}

	bts, err := json.Marshal(&c)
	if err != nil {
		return NodeConfiguration{}, err
	}

	res, err := n.api.Post("v1/node", "application/json", bytes.NewBuffer(bts))
	if err != nil {
		return NodeConfiguration{}, err
	}

	if res.StatusCode != http.StatusCreated {
		b := new(bytes.Buffer)
		io.Copy(b, res.Body)
		return NodeConfiguration{}, errors.New(b.String())
	}

	return c, nil
}
