package http

import (
	"github.com/alabianca/snfs/snfs"
	"net/http"
)

type Handler http.Handler

type AppHandler func(ctx *snfs.AppContext) Handler

type appHandler struct {
	appContext *snfs.AppContext
	handler http.Handler
}

func App(ctx *snfs.AppContext, handler AppHandler) http.Handler {
	return &appHandler{
		appContext: ctx,
		handler:    handler(ctx),
	}
}

func (a *appHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	a.handler.ServeHTTP(res, req)
}
