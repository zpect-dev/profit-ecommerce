package cart

import (
	"context"
	"time"
)

// CartItem representa un artículo individual dentro del carrito.
type CartItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Cart representa el carrito de compras de un usuario.
type Cart struct {
	UserID    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CatalogService define la comunicación externa para validar stocks.
type CatalogService interface {
	CheckStock(ctx context.Context, productIDs []string) (map[string]int, error)
}

// CartCacheRepository define las operaciones soportadas por la caché (Redis).
type CartCacheRepository interface {
	SaveCart(ctx context.Context, cart Cart) error
	GetCart(ctx context.Context, userID string) (Cart, error)
	DeleteCart(ctx context.Context, userID string) error
}

// CartDBRepository define las operaciones de persistencia permanente (Base de Datos).
type CartDBRepository interface {
	PersistCart(ctx context.Context, cart Cart) error
}
