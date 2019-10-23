package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type subscriptionRequest struct {
	Instance string `json:"instance"`
}

type subscriptionResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type MdnsService struct {
	api *RestAPI
}

func NewMdnsService() *MdnsService {
	return &MdnsService{
		api: NewRestAPI(),
	}
}

func (m *MdnsService) Register(instance string) (string, error) {
	bodyData, err := marshalSubscriptionRequest(instance)
	if err != nil {
		return "", err
	}

	return m.post("v1/mdns/subscribe", bytes.NewBuffer(bodyData))

}

func (m *MdnsService) Unregister() (string, error) {
	return m.post("v1/mdns/unsubscribe", nil)
}

func (m *MdnsService) post(url string, body io.Reader) (string, error) {
	res, err := m.api.Post(url, body)
	if err != nil {
		return "", err
	}

	var responseMessage subscriptionResponse
	if err := json.NewDecoder(res.Body).Decode(&responseMessage); err != nil {
		return "", err
	}

	if responseMessage.Status != http.StatusOK {
		return "Something went wrong", errors.New(responseMessage.Message)
	}

	return responseMessage.Message, nil
}

func marshalSubscriptionRequest(instance string) ([]byte, error) {
	req := subscriptionRequest{
		Instance: instance,
	}

	return json.Marshal(&req)
}
