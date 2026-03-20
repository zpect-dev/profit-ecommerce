package cart

import (
	"encoding/json"
	"net/http"

	"profit-ecommerce/internal/api/middleware"
)

// CartHandler expone la capa HTTP para el dominio del Carrito.
type CartHandler struct {
	service Service
}

// NewCartHandler inicializa los handlers inyectando el orquestador Service.
func NewCartHandler(service Service) *CartHandler {
	return &CartHandler{
		service: service,
	}
}

// HandleAddToCart maneja la petición POST para agregar ítems al carrito.
// Utiliza la librería estándar encoding/json y extrae el UserID del header.
func (h *CartHandler) HandleAddToCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var item CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.service.AddToCart(r.Context(), userID, item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// HandleGetCart maneja la petición GET para obtener el carrito habiéndolo validado previamente.
func (h *CartHandler) HandleGetCart(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	cart, err := h.service.GetValidatedCart(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cart); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
