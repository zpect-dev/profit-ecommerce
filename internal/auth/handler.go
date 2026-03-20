package auth

import (
	"encoding/json"
	"net/http"
	"strings"
)

// loginRequest representa el body JSON esperado en el endpoint de Login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthHandler expone la capa HTTP del dominio de autenticación.
type AuthHandler struct {
	svc AuthService
}

// NewAuthHandler crea un nuevo handler inyectando el servicio de Auth.
func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// HandleLogin procesa POST /auth/login.
// Decodifica las credenciales, llama al servicio y devuelve el perfil completo del cliente.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleLogout procesa POST /auth/logout.
// Extrae el token del header Authorization y lo elimina de Redis.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.Logout(r.Context(), token); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
