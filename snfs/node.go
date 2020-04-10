package snfs

type NodeConfiguration struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Cport     int    `json:"cport"`
	Dport     int    `json:"dport"`
	ProcessId int    `json:"processId"`
	NodeID    string `json:"nodeId"`
}
