package client

import (
	"github.com/alabianca/snfs/snfs/discovery"
	"github.com/alabianca/snfs/snfs/fs"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

func restAPIRoutes(c *ConnectivityService) *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		//setupCORS().Handler,        // Allow Cross-Origin-Requests
		middleware.Logger,          // Log API Requests
		middleware.DefaultCompress, // Compress results
		middleware.RedirectSlashes, // Redirect slashes to no slash url versions
		middleware.Recoverer,       // recover from panic without crashing
	)

	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/mdns", mdnsRoutes(c.discovery))
		r.Mount("/filesystem", filesystemRoutes(c.storage))
	})

	return router
}

func mdnsRoutes(d *discovery.Manager) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/subscribe", startMDNSController(d))
	router.Post("/unsubscribe", stopMDNSController(d))
	router.Get("/instance/{instance}", lookupMDNSController(d))

	return router
}

func filesystemRoutes(storage *fs.Manager) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/context", postFile(storage))

	return router
}
