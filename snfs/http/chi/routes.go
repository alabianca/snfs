package chi

import (
	"github.com/alabianca/snfs/snfs"
	"github.com/alabianca/snfs/snfs/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

)

func Routes(appContext *snfs.AppContext) http.Handler {
	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,          // Log API Requests
		middleware.RedirectSlashes, // Redirect slashes to no slash url versions
		middleware.Recoverer,       // recover from panic without crashing
	)

	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/storage", storageRoutes(appContext.Storage))
	})

	return router
}

func storageRoutes(storage snfs.Storage) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/fname/{name}", http.PostObjectController(storage))
	router.Get("/fname/{hash}", http.GetObjectController(storage))

	return router
}
