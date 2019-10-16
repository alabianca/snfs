package client

import "net"

type SubscribeRequest struct {
	Instance string `json:"instance"`
	Port     int    `json:"port"`
}

type LookupRequest struct {
	Instance string `json:"instance"`
}

type LookupResponse struct {
	IPs []net.IP `json:"ips"`
}

type FileUploadRequest struct {
	Path string `json:"path"`
}

type FileUploadResponse struct {
	Message string `json:"message"`
}
