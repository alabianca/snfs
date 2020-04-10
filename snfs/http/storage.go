package http

import (
	"github.com/alabianca/snfs/snfs"
	"net/http"
)

func PostObjectController(storage snfs.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
	}
}

func GetObjectController(storage snfs.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

	}
}
