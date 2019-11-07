package services

import (
	"bytes"
	"encoding/json"
)

type boostrapRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type BoostrapService struct {
	api *RestAPI
}

func NewBootstrapService() *BoostrapService {
	return &BoostrapService{
		api: NewRestAPI(),
	}
}

func (b *BoostrapService) Boostrap(port int, address, id string) error {
	req := boostrapRequest{
		ID:      id,
		Port:    port,
		Address: address,
	}

	bts, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	b.api.Post("v1/discovery/bootstrap", "application/json", bytes.NewBuffer(bts))

	return nil
}
