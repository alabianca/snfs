package http

import (
	"github.com/alabianca/snfs/snfsd"
	"net/http"
)

type AppHandler func(ctx *snfsd.AppContext) snfsd.Handler

type appHandler struct {
	appContext *snfsd.AppContext
	handler http.Handler
}

func App(appContext *snfsd.AppContext, handler AppHandler) http.Handler {
	h := handler(appContext)
	return &appHandler{
		appContext: appContext,
		handler:    h,
	}
}

func (h *appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
