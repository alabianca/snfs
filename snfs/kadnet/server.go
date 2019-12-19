package kadnet

import (
	"net"
	"strconv"

	"github.com/alabianca/gokad"
	"github.com/alabianca/snfs/snfs/kadnet/client"
	"github.com/alabianca/snfs/snfs/kadnet/conn"
	"github.com/alabianca/snfs/snfs/kadnet/kadmux"
	"github.com/alabianca/snfs/snfs/kadnet/messages"
	"github.com/alabianca/snfs/snfs/kadnet/request"
)

type Server struct {
	dht          *DHT
	host         string
	port         int
	mux          *kadmux.KadMux
	newClientReq chan *request.Request
}

func NewServer(dht *DHT, host string, port int) *Server {
	return &Server{
		dht:          dht,
		host:         host,
		port:         port,
		mux:          kadmux.NewMux(),
		newClientReq: make(chan *request.Request),
	}
}

func (s *Server) Listen() error {
	s.registerRequestHandlers()

	c, err := s.listen()
	if err != nil {
		return err
	}

	conn := conn.NewConn(c)
	nwf := conn.WriterFactory()
	defer conn.Close()

	go s.handleClientRequests(nwf)

	return s.mux.Start(conn, nwf)
}

func (s *Server) Shutdown() {
	s.mux.Shutdown()
}

func (s *Server) NewClient() *client.Client {
	return &client.Client{
		ID:    s.dht.Table.ID.String(),
		DoReq: s.newClientReq,
	}
}

func (s *Server) listen() (net.PacketConn, error) {

	conn, err := net.ListenPacket("udp", net.JoinHostPort(s.host, strconv.Itoa(s.port)))

	return conn, err
}

func (s *Server) registerRequestHandlers() {
	s.mux.HandleFunc(messages.FindNodeReq, s.onFindNode())
}

func (s *Server) handleClientRequests(nwf func(addr net.Addr) conn.KadWriter) {
	for req := range s.newClientReq {
		writer := nwf(req.Address())
		go s.doRequest(req, writer)
	}
}

func (s *Server) doRequest(req *request.Request, w conn.KadWriter) {
	data, err := req.Body.Bytes()
	if err != nil {
		return
	}

	w.Write(data)
}

// RPC Handlers
func (s *Server) onFindNode() kadmux.RpcHandler {
	return func(conn conn.KadWriter, req *request.Request) {
		fnr, ok := req.Body.(*messages.FindNodeRequest)
		if !ok {
			return
		}

		dht := GetDHT()
		id, err := gokad.From(fnr.Payload)
		if err != nil {
			return
		}

		contacts := dht.FindNode(id)

		res := messages.FindNodeResponse{
			SenderID:     dht.Table.ID.String(),
			EchoRandomID: fnr.RandomID,
			Payload:      contacts,
			RandomID:     gokad.GenerateRandomID().String(),
		}

		bts, err := res.Bytes()
		if err != nil {
			return
		}

		if _, err := conn.Write(bts); err != nil {
		}
	}
}
