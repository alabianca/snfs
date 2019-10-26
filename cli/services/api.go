package services

import (
	"io"
	"net/http"
)

type RestAPI struct {
	baseURL    string
	httpClient http.Client
}

func NewRestAPI() *RestAPI {
	return &RestAPI{
		baseURL:    "http://localhost:4200/api/",
		httpClient: http.Client{},
	}
}

func (r *RestAPI) Post(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", r.getURL(url), body)
	if err != nil {
		return nil, err
	}

	return r.httpClient.Do(req)

}

func (r *RestAPI) Get(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("GET", r.getURL(url), body)
	if err != nil {
		return nil, err
	}

	return r.httpClient.Do(req)
}

func (r *RestAPI) getURL(url string) string {
	return r.baseURL + url
}
