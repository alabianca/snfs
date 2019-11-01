package client

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/snfs/snfs/fs"

	"github.com/go-chi/chi"

	"github.com/alabianca/snfs/snfs/discovery"
)

func startMDNSController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, req.Body); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "MDNS Could Not Register"))
			return
		}

		var sReq SubscribeRequest
		if err := json.Unmarshal(buf.Bytes(), &sReq); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "MDNS Could Not Register"))
			return
		}

		if err := d.Register(sReq.Instance); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "MDNS Could Not Register"))
			return
		}

		util.Respond(res, util.Message(http.StatusOK, "MDNS Is Registered"))
	}
}

func stopMDNSController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		d.UnRegister()
		util.Respond(res, util.Message(http.StatusOK, "MDNS Stopped"))
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

func getInstancesController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		entries, err := d.Browse()
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Something went wrong"))
			return
		}

		if len(entries) == 0 {
			util.Respond(res, util.Message(http.StatusNotFound, "No Entrties found"))
			return
		}

		instances := make([]string, len(entries))

		for i, entry := range entries {
			instances[i] = entry.Instance
		}

		response := util.Message(http.StatusOK, "Ok")
		response["data"] = instances

		util.Respond(res, response)
	}
}

func storeFileController(storage *fs.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Println("Form Upload EP hit")
		req.ParseMultipartForm(10 << 20)

		file, header, err := req.FormFile("upload")
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, err.Error()))
			return
		}

		defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", header.Filename)
		fmt.Printf("File Size: %+v\n", header.Size)
		fmt.Printf("MIME Header: %+v\n", header.Header)

		destFile, err := fs.NewFile(storage.GetRoot(), header.Filename)
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Error Creating Destination File"))
			return
		}

		hasher := md5.New()
		storageWriter := fs.NewWriter(hasher, destFile)
		defer storageWriter.Close()

		if _, err := io.Copy(storageWriter, file); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Error Writing To Storage"))
			return
		}

		storageWriter.Close()

		hashed := fmt.Sprintf("%x", storageWriter.Sum(nil))
		if err := storage.AddObject(header.Filename, hashed, header.Size); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Error Adding File Object"))
			return
		}

		response := util.Message(http.StatusCreated, "OK")
		response["data"] = hashed

		util.Respond(res, response)

	}
}

func getObjectName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
