package network

// type Option func(n *Node)

// type Node struct {
// 	MaxTimeout   time.Duration
// 	Name         string
// 	Certfificate *tls.Certificate
// 	Port         int
// 	Host         string
// 	mdnsService  discovery.MDNS
// 	server       *server
// 	rootContext  string
// }

// func NewNode(o ...Option) *Node {
// 	n := Node{
// 		MaxTimeout: time.Second * 15,
// 		Host:       "127.0.0.1",
// 		Port:       4200,
// 	}

// 	for _, opt := range o {
// 		opt(&n)
// 	}

// 	return &n
// }

// func (n *Node) RegisterMdnsService(mdns discovery.MDNS) {
// 	n.mdnsService = mdns
// }

// // Bootstrap will register the mdns service and announce
// // it's presence to the local network. Boostrap
// // is also starting the local underlying tcp server
// func (n *Node) Bootstrap() error {
// 	if n.mdnsService == nil {
// 		return fmt.Errorf("Can't bootstrap with nil mdns service")
// 	}

// 	if err := n.mdnsService.Register(); err != nil {
// 		return err
// 	}

// 	n.server = &server{
// 		host:        n.Host,
// 		port:        n.Port,
// 		rootContext: new(bytes.Buffer),
// 	}

// 	return nil
// }

// func (n *Node) ListenAndServe() error {
// 	if n.server == nil {
// 		return fmt.Errorf("server must be initalized before serving")
// 	}

// 	return n.server.listen()
// }

// func (n *Node) Lookup(instance string) (*zeroconf.ServiceEntry, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), n.MaxTimeout)
// 	results := make(chan *zeroconf.ServiceEntry)
// 	n.mdnsService.Lookup(ctx, instance, results)

// 	defer cancel()
// 	var res *zeroconf.ServiceEntry
// 	for entry := range results {
// 		if entry.Instance == instance {
// 			res = entry
// 			cancel()
// 		}
// 	}

// 	if res == nil {
// 		return nil, fmt.Errorf("No Results Found")
// 	}

// 	return res, nil
// }

// func (n *Node) SetRootContext(path string) error {
// 	n.rootContext = path
// 	return util.WriteTarball(n.server.rootContext, path)
// }

// // Dial attempts to connect to instance
// // Dial will first do a lookup using mdns, the first successfull conn will be returned
// func (n *Node) Dial(instance string) (net.Conn, error) {
// 	service, err := n.Lookup(instance)
// 	if err != nil {
// 		return nil, err
// 	}

// 	port := service.Port
// 	var conn net.Conn
// 	var errc error
// 	for _, ip := range append(service.AddrIPv4, service.AddrIPv6...) {
// 		conn, errc = n.connect(port, ip)
// 		if errc == nil {
// 			break
// 		}
// 	}

// 	return conn, errc
// }

// func (n *Node) connect(port int, ip net.IP) (net.Conn, error) {
// 	p := strconv.Itoa(port)
// 	addr := net.JoinHostPort(ip.String(), p)

// 	return net.Dial("tcp", addr)
// }
