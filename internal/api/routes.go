package api

import (
	"net/http"

	"profit-ecommerce/internal/api/handlers"
	"profit-ecommerce/internal/api/middleware"
	"profit-ecommerce/internal/auth"
	"profit-ecommerce/internal/cart"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter crea el router HTTP. Recibe dependencias ya construidas (DI desde cmd/api/main.go).
func NewRouter(catHandler *handlers.CatalogHandler, cartHandler *cart.CartHandler, authHandler *auth.AuthHandler, authSvc auth.AuthService) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://192.168.4.217:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/v1", func(r chi.Router) {

		// --- Ruta Pública (sin autenticación) ---
		r.Post("/auth/login", authHandler.HandleLogin)

		// --- Rutas Protegidas (requieren Bearer token válido) ---
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(authSvc))

			r.Post("/auth/logout", authHandler.HandleLogout)

			r.Route("/products", func(r chi.Router) {
				r.Get("/", catHandler.List)
				r.Get("/{id}", catHandler.Single)
				r.Get("/categories", catHandler.Categories)
				r.Post("/batch", catHandler.GetByIDs)
			})

			r.Route("/cart", func(r chi.Router) {
				r.Get("/", cartHandler.HandleGetCart)
				r.Post("/", cartHandler.HandleAddToCart)
			})
		})
	})

	return r
}
