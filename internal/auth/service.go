package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type authService struct {
	cache    AuthCacheRepository
	clientDB ClientRepository
}

// NewAuthService inicializa el servicio de autenticación inyectando sus repositorios.
func NewAuthService(cache AuthCacheRepository, clientDB ClientRepository) AuthService {
	return &authService{cache: cache, clientDB: clientDB}
}

// Login valida las credenciales, busca al cliente en PostgreSQL, verifica que no esté inactivo,
// genera una sesión enriquecida con el perfil completo y la guarda en Redis.
func (s *authService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// 1. Validación básica de credenciales
	if username != password {
		return nil, errors.New("unauthorized")
	}

	// 2. Buscar cliente en PostgreSQL
	client, err := s.clientDB.FindClientByID(ctx, username)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	// 3. Verificar que el cliente no esté inactivo
	if client.Inactivo {
		return nil, errors.New("unauthorized: client inactive")
	}

	// 4. Generar token y crear sesión enriquecida
	token := uuid.New().String()

	session := Session{
		Token:     token,
		UserID:    client.CoCli,
		CliDes:    client.CliDes,
		Tipo:      client.Tipo,
		MontCre:   client.Login,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.cache.SaveSession(ctx, session); err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:   token,
		UserID:  client.CoCli,
		CliDes:  client.CliDes,
		Tipo:    client.Tipo,
		MontCre: client.Login,
	}, nil
}

// Logout cierra la sesión del usuario eliminando el correspondiente token de Redis.
func (s *authService) Logout(ctx context.Context, token string) error {
	return s.cache.DeleteSession(ctx, token)
}

// ValidateToken revisa si un token existe en la caché y si su marca de tiempo se encuentra
// completamente vigente. Omite devolver errores subyacentes con "unauthorized" por seguridad.
func (s *authService) ValidateToken(ctx context.Context, token string) (string, error) {
	session, err := s.cache.GetSession(ctx, token)
	if err != nil {
		return "", errors.New("unauthorized")
	}

	if time.Now().After(session.ExpiresAt) {
		return "", errors.New("unauthorized")
	}

	return session.UserID, nil
}
