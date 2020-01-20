package client

import (
	"github.com/alabianca/snfs/snfs/discovery"
	"github.com/alabianca/snfs/snfs/fs"
	"github.com/alabianca/snfs/snfs/kad"
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
		middleware.RedirectSlashes, // Redirect slashes to no slash url versions
		middleware.Recoverer,       // recover from panic without crashing
	)

	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/mdns", mdnsRoutes(c.discovery))
		r.Mount("/storage", storageRoutes(c.storage))
		r.Mount("/kad", kadnetRoutes(c.rpc))
	})

	return router
}

func mdnsRoutes(d *discovery.Manager) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/subscribe", startMDNSController(d))
	router.Post("/unsubscribe", stopMDNSController(d))
	router.Get("/instance/{instance}", lookupMDNSController(d))
	router.Get("/instance", getInstancesController(d))

	return router
}

func storageRoutes(storage *fs.Manager) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/fname/{name}", storeFileController(storage))
	router.Get("/fname/{hash}", getFileController(storage))

	return router
}

func kadnetRoutes(rpc *kad.RpcManager) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/bootstrap", bootstrapController(rpc))
	router.Get("/status", kadnetStatusController(rpc))

	return router
}
