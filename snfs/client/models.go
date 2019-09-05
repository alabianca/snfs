package client

import "net"

type LookupRequest struct {
	Instance string `json:"instance"`
}

type LookupResponse struct {
	IPs []net.IP `json:"ips"`
}
