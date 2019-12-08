package kadnet

import "net"

type Server struct {
	dht *DHT
	host string
	port int
	mux *KadMux
}

func NewServer(dht *DHT, host string, port int) *Server {
	return &Server{
		dht:  dht,
		host: host,
		port: port,
		mux:  NewMux(),
	}
}

func (s *Server) Listen() error {
	s.registerRequestHandlers()

	conn, err := s.listen()
	if err != nil {
		return err
	}

	return s.mux.start(conn)
}

func (s *Server) Shutdown() {
	s.mux.shutdown()
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
	return func(conn *net.UDPConn, req *Message) {

	}
}
