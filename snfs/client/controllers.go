package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/go-chi/chi"

	"github.com/alabianca/snfs/snfs/discovery"
)

func startMDNSController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, req.Body); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("MDNS Could Not Register"))
			return
		}

		var sReq SubscribeRequest
		if err := json.Unmarshal(buf.Bytes(), &sReq); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("MDNS Could Not Register"))
			return
		}

		if err := d.Register(sReq.Instance); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("MDNS Could Not Register"))
			return
		}
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte("MDNS Started"))
	}
}

func stopMDNSController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		d.UnRegister()
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte("MDNS Stopped"))
	}
}

func lookupMDNSController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		instance := chi.URLParam(req, "instance")
		if instance == "" {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Instance not provided"))
		}

		lookupReq := &LookupRequest{
			Instance: instance,
		}

		ips, err := d.Resolve(lookupReq.Instance)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte("Did not find ips for " + lookupReq.Instance))
			res.Write([]byte(err.Error()))
			return
		}

		resonse := &LookupResponse{
			IPs: ips,
		}

		out, err := json.Marshal(resonse)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Cannot marshall json"))
			return
		}

		res.WriteHeader(http.StatusOK)
		res.Header().Set("Content-Type", "application/json")
		res.Write(out)
	}
}

func postFile(manager *fs.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		buf := new(bytes.Buffer)
		var response FileUploadResponse
		var request FileUploadRequest
		if _, err := io.Copy(buf, req.Body); err != nil {
			res.WriteHeader(http.StatusBadRequest)
			response.Message = "Could Not Read Body"
			bts, _ := json.Marshal(&response)
			res.Write(bts)
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &request); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			response.Message = "Error Decoding Request " + err.Error()
			bts, _ := json.Marshal(&response)
			res.Write(bts)
			return
		}

		storage, err := fs.NewFile(getObjectName(request.Path))
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			response.Message = "Could Not Create Temporary File"
			bts, _ := json.Marshal(&response)
			res.Write(bts)
			return
		}
		defer storage.Close()

		gzw := gzip.NewWriter(storage)

		if err := fs.WriteTarball(gzw, request.Path); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			response.Message = "Error occured saving tarball"
			bts, _ := json.Marshal(&response)
			res.Write(bts)
			return
		}

		manager.Add(storage.Name())

		res.WriteHeader(http.StatusCreated)
		response.Message = "File " + storage.Name() + " created"
		bts, _ := json.Marshal(&response)
		res.Write(bts)
	}
}

func getObjectName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
