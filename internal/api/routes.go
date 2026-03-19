package api

import (
	"net/http"
	"profit-ecommerce/internal/api/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter crea el router HTTP. Recibe dependencias ya construidas (DI desde cmd/api/main.go).
func NewRouter(catHandler *handlers.CatalogHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://192.168.4.64:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/v1", func(r chi.Router) {

		r.Route("/products", func(r chi.Router) {
			r.Get("/", catHandler.List)

			r.Get("/{id}", catHandler.Single)

			r.Get("/categories", catHandler.Categories)

			r.Post("/batch", catHandler.GetByIDs)
			//r.Get("/categories/{id}", catHandler.ListByCategory)
		})
	})

	return r
}
