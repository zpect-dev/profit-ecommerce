package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisAuthRepository implementa AuthCacheRepository sirviéndose del driver go-redis.
type redisAuthRepository struct {
	client *redis.Client
}

// NewRedisAuthRepository crea el repositorio para auth inyectando la dependencia de Redis explícita.
func NewRedisAuthRepository(client *redis.Client) AuthCacheRepository {
	return &redisAuthRepository{
		client: client,
	}
}

// SaveSession serializa la sesión proveída y la deposita en el backend key-value con un TTL directo en base de datos.
func (r *redisAuthRepository) SaveSession(ctx context.Context, session Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := "session:" + session.Token

	// Garantizamos un TTL de 24 horas o el cálculo restante para que la DB se limpie estricta y autónomamente.
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetSession busca un token particular en forma veloz desde caché y lo deserializa.
func (r *redisAuthRepository) GetSession(ctx context.Context, token string) (Session, error) {
	key := "session:" + token

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return Session{}, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, err
	}

	return session, nil
}

// DeleteSession erradica la clave y su valor adjunto instantáneamente.
func (r *redisAuthRepository) DeleteSession(ctx context.Context, token string) error {
	key := "session:" + token
	return r.client.Del(ctx, key).Err()
}
