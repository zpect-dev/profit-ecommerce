package auth

import (
	"context"
	"time"
)

// Session representa una sesión de usuario almacenada en caché (Redis).
// Contiene el perfil completo del cliente para que el Frontend lo consuma al loguearse.
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	CliDes    string    `json:"cli_des"`
	Tipo      string    `json:"tipo"`
	MontCre   float64   `json:"mont_cre"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginResponse es la estructura que se devuelve al Frontend tras un login exitoso.
// Excluye ExpiresAt interna para no exponer detalles de implementación.
type LoginResponse struct {
	Token   string  `json:"token"`
	UserID  string  `json:"user_id"`
	CliDes  string  `json:"cli_des"`
	Tipo    string  `json:"tipo"`
	MontCre float64 `json:"mont_cre"`
}

// ClientRow representa una fila de la tabla clientes en PostgreSQL para el lookup de Login.
type ClientRow struct {
	CoCli    string  `db:"co_cli"`
	Tipo     string  `db:"tipo"`
	CliDes   string  `db:"cli_des"`
	Inactivo bool    `db:"inactivo"`
	Login    float64 `db:"login"`
}

// ClientRepository define el contrato de lectura contra PostgreSQL para autenticación.
type ClientRepository interface {
	FindClientByID(ctx context.Context, coCli string) (ClientRow, error)
}

// AuthCacheRepository define la persistencia en caché de las sesiones usando Redis.
type AuthCacheRepository interface {
	SaveSession(ctx context.Context, session Session) error
	GetSession(ctx context.Context, token string) (Session, error)
	DeleteSession(ctx context.Context, token string) error
}

// AuthService define la orquestación y las reglas de negocio de autenticación.
type AuthService interface {
	Login(ctx context.Context, username, password string) (*LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (string, error)
}
