package cart

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisCartRepository implementa CartCacheRepository usando go-redis.
type redisCartRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCartRepository crea una nueva instancia del repositorio de caché del carrito.
func NewRedisCartRepository(client *redis.Client) CartCacheRepository {
	return &redisCartRepository{
		client: client,
		ttl:    24 * time.Hour, // Tiempo de vida del carrito en Redis
	}
}

// SaveCart serializa el carrito y lo guarda en Redis usando un Hash.
// Debido a que se necesitan operaciones atómicas en rediseño de alta concurrencia, usamos TxPipeline.
func (r *redisCartRepository) SaveCart(ctx context.Context, cart Cart) error {
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}

	key := r.getCartKey(cart.UserID)

	// Regla Estricta de Performance: TxPipeline para agrupar comandos en una sola operación atómica.
	pipe := r.client.TxPipeline()

	// Usamos HSet para guardar el payload JSON
	pipe.HSet(ctx, key, "data", data)
	// Establecemos el TTL en el mismo pipeline
	pipe.Expire(ctx, key, r.ttl)

	// Ejecutamos ambas sin realizar múltiples operaciones de red asíncronas
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetCart obtiene un carrito desde Redis basado en el userID.
func (r *redisCartRepository) GetCart(ctx context.Context, userID string) (Cart, error) {
	key := r.getCartKey(userID)

	data, err := r.client.HGet(ctx, key, "data").Bytes()
	if err != nil {
		return Cart{}, err
	}

	var cart Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return Cart{}, err
	}

	return cart, nil
}

// DeleteCart elimina un carrito de Redis.
func (r *redisCartRepository) DeleteCart(ctx context.Context, userID string) error {
	key := r.getCartKey(userID)
	return r.client.Del(ctx, key).Err()
}

// getCartKey encapsula la convención de nombres para las claves en Redis.
func (r *redisCartRepository) getCartKey(userID string) string {
	return "cart:" + userID
}
