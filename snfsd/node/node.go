package node

import (
	"bytes"
	"github.com/alabianca/snfs/snfsd"
)

type NodeService struct {
	Node snfsd.Node
	Publisher snfsd.Publisher
}

func (n *NodeService) Create(nc *snfsd.NodeConfiguration) error {
	if err := n.Node.Create(nc); err != nil {
		return err
	}

	return n.publish(nc)
}

func (n *NodeService) Delete(id int) error {
	if err := n.Node.Delete(id); err != nil {
		return err
	}
	
	nc := snfsd.NodeConfiguration{
		ProcessId: id,
	}

	return n.publish(&nc)
	
}

func (n *NodeService) publish(data interface{}) error {
	buf := new(bytes.Buffer)
	if err := snfsd.Encode(buf, data); err != nil {
		return err
	}

	n.Publisher.Publish(snfsd.TopicAddNode, buf.Bytes())

	return nil
}
