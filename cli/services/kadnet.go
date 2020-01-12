package services

import (
	"bytes"
	"encoding/json"
	"github.com/alabianca/gokad"
)

type boostrapRequest struct {
	ID      string `json:"id"`
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

func NewKadnetService() *KadnetService {
	return &KadnetService{
		api: NewRestAPI(getBaseURL()),
	}
}

func (k *KadnetService) Boostrap(port int, address, id string) error {
	req := boostrapRequest{
		ID:      id,
		Port:    port,
		Address: address,
	}

	bts, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	k.api.Post("v1/kadnet/bootstrap", "application/json", bytes.NewBuffer(bts))

	return nil
}

func (k *KadnetService) GetStatus() ([]RoutingTableEntry, error) {
	res, err := k.api.Get("v1/kadnet/status", nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	var data KadnetStatusResponse
	if err := decode(res.Body, &data); err != nil {
		return nil, err
	}

	return data.Entries, nil
}
