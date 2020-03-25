package snfsd

type Node interface {
	Create(nc *NodeConfiguration) error
	Delete(id int) error
}
