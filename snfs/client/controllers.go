package client

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/alabianca/snfs/snfs/server"

	"github.com/grandcat/zeroconf"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/snfs/snfs/fs"
	"github.com/alabianca/snfs/snfs/kad"

	"github.com/go-chi/chi"

	"github.com/alabianca/snfs/snfs/discovery"
)

func startMDNSController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if d.MDNSStarted() {
			util.Respond(res, util.Message(http.StatusBadRequest, "MDNS Already Started"))
			return
		}

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

		d.SetInstance(sReq.Instance)

		mdns := make(chan server.ResponseCode, 1)
		rpc := make(chan server.ResponseCode, 1)

		if err := queueServiceRequest(discovery.ServiceName, server.OPStartService, mdns); err != nil {
			util.Respond(res, util.Message(http.StatusNotFound, "Could Not Resolve Service "+discovery.ServiceName))
			return
		}
		if err := queueServiceRequest(kad.ServiceName, server.OPStartService, rpc); err != nil {
			util.Respond(res, util.Message(http.StatusNotFound, "Could Not Resolve Service "+discovery.ServiceName))
			return
		}

		<-mdns
		<-rpc

		util.Respond(res, util.Message(http.StatusOK, "MDNS Is Registered"))
	}
}

func stopMDNSController(d *discovery.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		mdns := make(chan server.ResponseCode, 1)
		rpc := make(chan server.ResponseCode, 1)
		if err := queueServiceRequest(discovery.ServiceName, server.OPStopService, mdns); err != nil {
			util.Respond(res, util.Message(http.StatusNotFound, "Could Not Resolve Service "+discovery.ServiceName))
			return
		}
		if err := queueServiceRequest(kad.ServiceName, server.OPStopService, rpc); err != nil {
			util.Respond(res, util.Message(http.StatusNotFound, "Could Not Resolve Service "+discovery.ServiceName))
			return
		}

		<-mdns
		<-rpc

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

		instances := make([]InstanceResponse, len(entries))

		mapper := func(entry *zeroconf.ServiceEntry) InstanceResponse {
			var instance InstanceResponse
			instance.InstanceName = entry.Instance

			for _, txt := range entry.Text {
				keyVal := strings.Split(txt, ":")
				key := keyVal[0]
				val := keyVal[1]
				if key == "NodeID" {
					instance.ID = val
				}
				if key == "Port" {
					if p, err := strconv.ParseInt(val, 10, 16); err == nil {
						instance.Port = p
					}
				}

				if key == "Address" {
					instance.Address = val
				}
			}

			return instance
		}

		for i, entry := range entries {
			instances[i] = mapper(entry)
		}

		response := util.Message(http.StatusOK, "Ok")
		response["data"] = instances

		util.Respond(res, response)
	}
}

func storeFileController(storage *fs.Manager, rpc *kad.RpcManager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		req.ParseMultipartForm(100 << 20) // 100mgb

		file, header, err := req.FormFile("upload")
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, err.Error()))
			return
		}

		defer file.Close()

		destFile, err := fs.NewFile(storage.GetRoot(), header.Filename)
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Error Creating Destination File"))
			return
		}

		hasher := sha1.New()
		storageWriter := fs.NewWriter(hasher, destFile)
		defer storageWriter.Close()

		var bytesWritten int64
		if bytesWritten, err = io.Copy(storageWriter, file); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Error Writing To Storage"))
			return
		}


		storageWriter.Close()

		hashed := fmt.Sprintf("%x", storageWriter.Sum(nil))
		if err := storage.AddObject(header.Filename, hashed, header.Size); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Error Adding File Object"))
			return
		}

		host := os.Getenv("SNFS_HOST")
		if host == "" {
			util.Respond(res, util.Message(http.StatusInternalServerError, "SNFS_HOST Environment Variable Not Set"))
			return
		}
		port, err := strconv.ParseInt(os.Getenv("SNFS_FS_PORT"), 10, 16)
		if err != nil || port == 0 {
			util.Respond(res, util.Message(http.StatusInternalServerError, "SNFS_FS_PORT Environment Variable Not Set"))
			return
		}


		if _, err := rpc.Store(hashed, net.ParseIP(host), int(port)); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, err.Error()))
			return
		}

		response := util.Message(http.StatusCreated, "OK")
		response["data"] = StorageResponse{
			Hash:        hashed,
			ByteWritten: bytesWritten,
		}

		util.Respond(res, response)

	}
}

func getFileController(storage *fs.Manager, rpc *kad.RpcManager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fileHash := chi.URLParam(req, "hash")

		addr, err := rpc.Resolve(fileHash)
		log.Printf("Resolved Address %s\n", addr)
		log.Printf("Resolved error %s\n", err)

		url := "http://" + addr.String() + "/v1/object/" + fileHash
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Could Not Create a Request"))
			return
		}

		client := http.Client{}
		response, err := client.Do(request)
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Could Not Do Request"))
			return
		}

		defer response.Body.Close()
		res.Header().Add("Content-Type", "application/octet-stream")
		res.WriteHeader(http.StatusOK)
		io.Copy(res, response.Body)
	}
}

func bootstrapController(rpc *kad.RpcManager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var br BootstrapRequest

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, req.Body); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, err.Error()))
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &br); err != nil {
			util.Respond(res, util.Message(http.StatusBadRequest, err.Error()))
			return
		}

		rpc.Bootstrap(br.Port, br.Address)

	}
}

func kadnetStatusController(rpc *kad.RpcManager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		sr := rpc.Status()

		response := util.Message(http.StatusOK, "Ok")
		response["data"] = sr
		util.Respond(res, response)


	}
}

func queueServiceRequest(serviceName string, op server.OP, res chan server.ResponseCode) error {
	service, err := server.ResolveService(serviceName)
	if err != nil {
		return err
	}

	req := server.ServiceRequest{
		Op:      op,
		Service: service,
		Res:     res,
	}

	server.QueueServiceRequest(req)

	return nil
}
