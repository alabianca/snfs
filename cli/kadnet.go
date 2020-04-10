package cli

import (
	"bytes"
	"encoding/json"
)

func NewKadnetService() *KadnetService {
	return &KadnetService{
		api: NewRestAPI(getBaseURL()),
	}
}

func (k *KadnetService) Boostrap(port int, address string) error {
	req := boostrapRequest{
		Port:    port,
		Address: address,
	}

	bts, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	k.api.Post("v1/kad/bootstrap", "application/json", bytes.NewBuffer(bts))

	return nil
}

func (k *KadnetService) GetStatus() ([]RoutingTableEntry, error) {
	res, err := k.api.Get("v1/kad/status", nil)
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
