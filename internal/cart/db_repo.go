package cart

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

type sqlxCartRepository struct {
	db *sqlx.DB
}

// NewSQLCartRepository crea una nueva instancia del Repositorio de Base de Datos para Carts usando sqlx.
func NewSQLCartRepository(db *sqlx.DB) CartDBRepository {
	return &sqlxCartRepository{db: db}
}

// PersistCart guarda el estado completo del carrito utilizando una estrategia UPSERT en PostgreSQL.
// Serializa los items como JSONB y actualiza las columnas si el user_id ya existe.
func (r *sqlxCartRepository) PersistCart(ctx context.Context, cart Cart) error {
	itemsJSON, err := json.Marshal(cart.Items)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO carts (user_id, items, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			items = EXCLUDED.items,
			updated_at = EXCLUDED.updated_at;
	`

	_, err = r.db.ExecContext(ctx, query, cart.UserID, itemsJSON, time.Now())
	return err
}
