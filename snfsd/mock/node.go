package mock

import "github.com/alabianca/snfs/snfsd"

type NodeService struct {
	CreateInvoked bool
	DeleteInvoked bool
	UpdateInvoked bool
	CreateFn func(n *snfsd.NodeConfiguration) error
	UpdateFn func(n snfsd.NodeConfiguration) error
	DeleteFn func(id int) error
}

func (n *NodeService) Create(nc *snfsd.NodeConfiguration) error {
	n.CreateInvoked = true
	return n.CreateFn(nc)
}

func (n *NodeService) Update(nc snfsd.NodeConfiguration) error {
	n.UpdateInvoked = true
	return n.UpdateFn(nc)
}

func (n *NodeService) Delete(id int) error {
	n.DeleteInvoked = true
	return n.DeleteFn(id)
}
