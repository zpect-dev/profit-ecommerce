package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Repository es la implementación concreta de CatalogRepository contra PostgreSQL.
type Repository struct {
	db *sqlx.DB
}

// NewRepository crea una nueva instancia del repositorio de catálogo.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error) {
	if len(ids) == 0 {
		return []Product{}, nil
	}

	baseQuery := getBaseProductQuery()
	query := fmt.Sprintf("%s WHERE a.co_art IN (?)", baseQuery)

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)

	var products []Product
	err = r.db.SelectContext(ctx, &products, query, args...)
	if err != nil {
		return nil, err
	}

	for i := range products {
		if len(products[i].InventarioRaw) > 0 {
			_ = json.Unmarshal(products[i].InventarioRaw, &products[i].Inventario)
		}
	}

	return products, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Product, error) {
	baseQuery := getBaseProductQuery()
	query := fmt.Sprintf("%s WHERE a.co_art = $1", baseQuery)

	var product Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		return Product{}, err
	}

	if len(product.InventarioRaw) > 0 {
		_ = json.Unmarshal(product.InventarioRaw, &product.Inventario)
	}

	return product, nil
}

func (r *Repository) ListProducts(ctx context.Context, page int, limit int, search string, category string, inStock, hasDiscount bool) ([]Product, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	baseQuery := getBaseProductQuery()

	// Filtros dinámicos
	extraFilters := ""
	if inStock {
		extraFilters += " AND a.stock_act > 0"
	}
	if hasDiscount {
		extraFilters += " AND (COALESCE(d_art.porc1, 0) > 0 OR COALESCE(d_cat.porc1, 0) > 0 OR COALESCE(d_lin.porc1, 0) > 0)"
	}

	query := fmt.Sprintf(`
        %s 
        WHERE (a.art_des ILIKE $1 OR a.co_art ILIKE $1 OR a.campo4 ILIKE $1) 
        AND ($4 = '' OR a.co_lin = $4)
        %s
        ORDER BY a.art_des ASC
        LIMIT $2 OFFSET $3
    `, baseQuery, extraFilters)

	searchTerm := "%" + search + "%"

	var products []Product
	err := r.db.SelectContext(ctx, &products, query, searchTerm, limit, offset, category)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo productos: %w", err)
	}

	for i := range products {
		if len(products[i].InventarioRaw) > 0 {
			_ = json.Unmarshal(products[i].InventarioRaw, &products[i].Inventario)
		}
	}

	return products, nil
}

func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	query := `SELECT co_lin, lin_des FROM lin_art WHERE co_lin NOT IN ('06', '07', '11', '12', '13', '14', '315')`

	var categories []Category
	err := r.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("error listando categorias: %w", err)
	}

	return categories, nil
}

func getBaseProductQuery() string {
	return `
       SELECT 
             a.co_art, 
             a.art_des,
             a.stock_act, 
             a.prec_vta1,
             a.prec_vta2,
             a.prec_vta3,
             a.prec_vta4,
             a.prec_vta5,
             a.tipo_imp,
             a.co_lin,
             a.co_cat,
             a.co_subl,
             COALESCE(a.image_url, '') AS image_url,
             COALESCE(a.inventory_json, '{}') AS inventario_detallado,
             COALESCE(d_art.porc1, 0) AS desc_articulo,
             COALESCE(d_cat.porc1, 0) AS desc_categoria,
             COALESCE(d_lin.porc1, 0) AS desc_linea
        FROM art a
        LEFT JOIN LATERAL (
            SELECT porc1 FROM descuen WHERE co_desc = a.co_art LIMIT 1
        ) d_art ON true
        LEFT JOIN LATERAL (
            SELECT porc1 FROM descuen WHERE co_desc = a.co_cat LIMIT 1
        ) d_cat ON true
        LEFT JOIN LATERAL (
            SELECT porc1 FROM descuen WHERE co_desc = a.co_lin LIMIT 1
        ) d_lin ON true
    `
}
