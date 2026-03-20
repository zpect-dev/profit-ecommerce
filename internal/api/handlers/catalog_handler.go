package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"profit-ecommerce/internal/catalog"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	repo catalog.CatalogRepository
}

func NewCatalogHandler(repo catalog.CatalogRepository) *CatalogHandler {
	return &CatalogHandler{repo: repo}
}

func (h *CatalogHandler) GetByIDs(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Formato de JSON invalido", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []interface{}{},
		})
		return
	}

	products, err := h.repo.GetProductsByIDs(r.Context(), req.IDs)
	if err != nil {
		fmt.Printf("Error obteniendo productos batch: %v\n", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  products,
		"count": len(products),
	})
}

func (h *CatalogHandler) Single(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "ID es requerido", http.StatusBadRequest)
		return
	}

	product, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		fmt.Printf("Error: %s", err)
		http.Error(w, "Erro interno en el servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": product,
	})
}

func (h *CatalogHandler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	category := r.URL.Query().Get("category")
	inStock := r.URL.Query().Get("in_stock") == "true" || r.URL.Query().Get("in_stock") == "1"
	hasDiscount := r.URL.Query().Get("has_discount") == "true" || r.URL.Query().Get("has_discount") == "1"

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	products, err := h.repo.ListProducts(r.Context(), page, limit, search, category, inStock, hasDiscount)
	if err != nil {
		fmt.Printf("Error ListProducts: %v\n", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  products,
		"page":  page,
		"limit": limit,
		"total": len(products),
	})
}

func (h *CatalogHandler) Categories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.ListCategories(r.Context())
	if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    categories,
		"success": true,
	})
}
