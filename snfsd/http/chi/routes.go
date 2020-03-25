package chi

import (
	"github.com/alabianca/snfs/snfsd"
	"github.com/alabianca/snfs/snfsd/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

func Routes(appContext *snfsd.AppContext) snfsd.Handler {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		setupCORS().Handler,
	)

	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/node", http.PostNodeController(appContext.NodeService))
	})

	return router
}

func setupCORS() *cors.Cors {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "GET", "UPDATE", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-TOKEN"},
		AllowCredentials: true,
		MaxAge:           500,
	})

	return c
}
