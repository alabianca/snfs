package peer

import (
	"github.com/alabianca/snfs/snfs/fs"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

func restAPIRoutes(m *Manager) *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		//setupCORS().Handler,        // Allow Cross-Origin-Requests
		middleware.Logger,          // Log API Requests
		middleware.DefaultCompress, // Compress results
		middleware.RedirectSlashes, // Redirect slashes to no slash url versions
		middleware.Recoverer,       // recover from panic without crashing
	)

	router.Route("/peer/v1", func(r chi.Router) {
		r.Mount("/resource", resourceRoutes(m.storage))
	})

	return router
}

func resourceRoutes(storage *fs.Manager) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/{resourceId}", getResource(storage))

	return router
}
