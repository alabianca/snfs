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

type MdnsService struct {
	api *RestAPI
}

func NewMdnsService() *MdnsService {
	return &MdnsService{
		api: NewRestAPI(getBaseURL()),
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

func (m *MdnsService) Browse() ([]Node, error) {
	res, err := m.api.Get("v1/mdns/instance", nil)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var responseMessage instancesResponse
	if err := decode(res.Body, &responseMessage); err != nil {
		return nil, err
	}

	if responseMessage.Status != http.StatusOK {
		return nil, errors.New(responseMessage.Message)
	}

	return responseMessage.Instances, nil
}

func (m *MdnsService) post(url string, body io.Reader) (string, error) {
	res, err := m.api.Post(url, "application/json", body)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var responseMessage subscriptionResponse
	if err := decode(res.Body, &responseMessage); err != nil {
		return "", nil
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
