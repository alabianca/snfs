package kadnet

type RPC interface {
	//NodeLookup()
	Bootstrap(port int, ip, idHex string)
	GetID() []byte
}

type rpcManager struct {
	dht     *DHT
	port    int
	address string
}

func NewRPCManager(address string, port int) RPC {
	return &rpcManager{
		dht:     NewDHT(),
		port:    port,
		address: address,
	}
}

func (rpc *rpcManager) Bootstrap(port int, ip, idHex string) {
	rpc.dht.Bootstrap(port, ip, idHex)
}

func (rpc *rpcManager) GetID() []byte {
	return rpc.dht.Table.ID.Bytes()
}
