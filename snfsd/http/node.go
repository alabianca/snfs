package http

import (
	"github.com/alabianca/snfs/snfsd"
	"github.com/alabianca/snfs/util"
	"net/http"
)

func PostNodeController(service snfsd.NodeService) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var body snfsd.NodeConfiguration
		if err := snfsd.Decode(req.Body, &body); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, "Error Decoding Request Body"))
			return
		}

		// @todo do validation here

		if err := service.Create(&body); err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, err.Error()))
			return
		}


		util.Respond(res, util.Message(http.StatusCreated, "OK"))
	}
}
