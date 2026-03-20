package auth

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// sqlClientRepository implementa ClientRepository contra PostgreSQL usando sqlx.
type sqlClientRepository struct {
	db *sqlx.DB
}

// NewClientRepository crea el repositorio de clientes inyectando la conexión a la BD.
func NewClientRepository(db *sqlx.DB) ClientRepository {
	return &sqlClientRepository{db: db}
}

// FindClientByID busca un cliente por su código principal (co_cli) en la tabla clientes.
func (r *sqlClientRepository) FindClientByID(ctx context.Context, coCli string) (ClientRow, error) {
	var client ClientRow
	query := `SELECT co_cli, tipo, cli_des, inactivo, login FROM clientes WHERE co_cli = $1`
	err := r.db.GetContext(ctx, &client, query, coCli)
	if err != nil {
		return ClientRow{}, fmt.Errorf("client not found: %w", err)
	}
	return client, nil
}
