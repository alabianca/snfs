package fs

import (
	"github.com/alabianca/snfs/util"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
)

type server struct {
	addr string
	port int
	server *http.Server
}

func (s *server) listen(fs *Manager) error {
	addr := net.JoinHostPort(s.addr, strconv.Itoa(s.port))
	s.server = &http.Server{
		Addr:              addr,
		Handler:           routes(fs),
		//TLSConfig:         nil,
		//ReadTimeout:       0,
		//ReadHeaderTimeout: 0,
		//WriteTimeout:      0,
		//IdleTimeout:       0,
		//MaxHeaderBytes:    0,
		//TLSNextProto:      nil,
		//ConnState:         nil,
		//ErrorLog:          nil,
	}

	return s.server.ListenAndServe()
}

func routes(fs *Manager) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Get("/v1/object/{hash}", getFile(fs))

	return router
}

func getFile(fs *Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		hash := chi.URLParam(req, "hash")
		if hash == "" {
			util.Respond(res, util.Message(http.StatusBadRequest,"File Hash Is Required"))
			return
		}

		filePath, err := fs.GetObjectPath(hash)
		if err != nil {
			util.Respond(res, util.Message(http.StatusNotFound, "File Not Found"))
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			util.Respond(res, util.Message(http.StatusInternalServerError, err.Error()))
			return
		}

		defer file.Close()

		res.WriteHeader(http.StatusOK)
		res.Header().Add("Content-Type", "application/octet-stream")
		io.Copy(res, file)
	}
}
