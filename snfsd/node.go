package snfsd

type NodeConfiguration struct {
	Name      string `json:"name"`
	Cport     int    `json:"cport"`
	Dport     int    `json:"dport"`
	Fport     int    `json:"fport"`
	ProcessId int    `json:"processId"`
}

type NodeService interface {
	Create(n *NodeConfiguration) error
	Delete(id int) error
}
