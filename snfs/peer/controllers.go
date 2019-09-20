package peer

import (
	"io"
	"net/http"

	"github.com/alabianca/snfs/snfs/fs"
	"github.com/go-chi/chi"
)

func getResource(storage *fs.Manager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")
		if id == "" {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Missing ID Param"))
			return
		}

		file, err := storage.Find(id)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte("File Not Found"))
			return
		}

		res.WriteHeader(http.StatusFound)
		io.Copy(res, file)
	}
}
