package client

import "net"

type SubscribeRequest struct {
	Instance string `json:"instance"`
}

type SubscribeRequestMessage struct {
	Message string `json:"message"`
}

type LookupRequest struct {
	Instance string `json:"instance"`
}

type LookupResponse struct {
	IPs []net.IP `json:"ips"`
}

type InstanceResponse struct {
	InstanceName string `json:"name"`
	ID           string `json:"id"`
	Port         int64  `json:"port"`
	Address      string `json:"address"`
}

type FileUploadRequest struct {
	Path string `json:"path"`
}

type FileUploadResponse struct {
	Message string `json:"message"`
}
