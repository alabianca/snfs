package kadnet

import "net"

type Server struct {
	dht *DHT
	host string
	port int
	mux *KadMux
	newClientReq chan *Request
}

func NewServer(dht *DHT, host string, port int) *Server {
	return &Server{
		dht:  dht,
		host: host,
		port: port,
		mux:  NewMux(),
		newClientReq: make(chan *Request),
	}
}

func (s *Server) Listen() error {
	s.registerRequestHandlers()

	c, err := s.listen()
	if err != nil {
		return err
	}

	conn := NewConn(c)
	defer conn.Close()

	return s.mux.start(conn)
}

func (s *Server) Shutdown() {
	s.mux.shutdown()
}

func (s *Server) NewClient() *Client {
	return &Client{
		id: s.dht.Table.ID.String(),
		doReq: s.newClientReq,
	}
}

func (s *Server) listen() (*net.UDPConn, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(s.host),
		Port: s.port,
	})

	return conn, err
}

func (s *Server) registerRequestHandlers() {
	s.mux.HandleFunc(FindNodeReq, s.onFindNode())
}


// RPC Handlers
func (s *Server) onFindNode() RpcHandler {
	return func(conn KadWriter, req *Request) {

	}
}
