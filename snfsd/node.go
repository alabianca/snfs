package snfsd

type NodeConfiguration struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	Cport     int    `json:"cport"`
	Dport     int    `json:"dport"`
	ProcessId int    `json:"processId,omitempty"`
	Started   int64  `json:"started,omitempty"`
	NodeId    string `json:"nodeId,omitempty"`
}

type NodeService interface {
	Create(n *NodeConfiguration) error
	Delete(id int) error
	Update(n NodeConfiguration) error
}
