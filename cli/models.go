package cli

import (
	"github.com/alabianca/gokad"
	"time"
)

type boostrapRequest struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type KadnetStatusResponse struct {
	Status  int                 `json:"status"`
	Message string              `json"message"`
	Entries []RoutingTableEntry `json:"data"`
}

type RoutingTableEntry struct {
	BucketIndex int        `json:"bucketIndex"`
	Contact     gokad.Contact `json:"contact"`
}

type KadnetService struct {
	api *RestAPI
}

type subscriptionRequest struct {
	Instance string `json:"instance"`
}

type subscriptionResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type Node struct {
	InstanceName string `json:"name"`
	ID           string `json:"id"`
	Port         int64  `json:"port"`
	Address      string `json:"address"`
}

type instancesResponse struct {
	Instances []Node `json:"data"`
	Status    int    `json:"status"`
	Message   string `json:"message"`
}

type storageResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Content StoreResult `json:"data"`
}

type StoreResult struct {
	Hash         string `json:"hash"`
	BytesWritten int64  `json:"bytesWritten"`
	Took         time.Duration
}

type NodeConfiguration struct {
	Name      string `json:"name"`
	Cport     int    `json:"cport"`
	Dport     int    `json:"dport"`
	ProcessId int    `json:"processId"`
}
