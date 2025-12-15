package catalog

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type AlmacenResumen struct {
	Nombre            string  `json:"nombre"`
	StockTotal        float64 `json:"stock_total"`
	StockComprometido float64 `json:"stock_comprometido"`
	StockPorLlegar    float64 `json:"stock_por_llegar"`
}

type Product struct {
	CoArt    string  `json:"co_art"`
	ArtDes   string  `json:"art_des"`
	StockAct float64 `json:"stock_act"`
	PrecVta1 float64 `json:"prec_vta1"`
	PrecVta2 float64 `json:"prec_vta2"`
	PrecVta3 float64 `json:"prec_vta3"`
	PrecVta4 float64 `json:"prec_vta4"`
	PrecVta5 float64 `json:"prec_vta5"`
	TipoImp  string  `json:"tipo_imp"`
	CoLin    string  `json:"co_lin"`
	CoCat    string  `json:"co_cat"`
	CoSubl   string  `json:"co_subl"`
	ImageUrl string  `json:"image_url"`

	// Este campo se llena leyendo la columna pre-calculada 'inventory_json'
	InventarioRaw []byte                    `json:"-"`
	Inventario    map[string]AlmacenResumen `json:"inventario"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByID(id string) (Product, error) {
	// OPTIMIZACIÓN: Eliminamos los JOINs y GROUP BY.
	// Ahora leemos directo la columna 'inventory_json' que el Syncer mantiene actualizada.
	query := `
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
             COALESCE(a.inventory_json, '{}') AS inventario_detallado -- ¡Lectura directa y rápida!
          FROM art a
          WHERE a.co_art = $1
       `

	product := Product{}

	err := r.db.QueryRow(query, id).Scan(
		&product.CoArt, &product.ArtDes, &product.StockAct,
		&product.PrecVta1, &product.PrecVta2, &product.PrecVta3, &product.PrecVta4, &product.PrecVta5,
		&product.TipoImp, &product.CoLin, &product.CoCat, &product.CoSubl,
		&product.ImageUrl, &product.InventarioRaw,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, fmt.Errorf("producto con id %s no encontrado", id)
		}
		return Product{}, fmt.Errorf("error en la base de datos: %w", err)
	}

	// Deserializamos el JSON rápido en memoria
	if len(product.InventarioRaw) > 0 {
		_ = json.Unmarshal(product.InventarioRaw, &product.Inventario)
	}

	return product, nil
}

func (r *Repository) ListProducts(page, limit int, search string) ([]Product, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	// OPTIMIZACIÓN: Misma lógica aquí. Leemos la columna ya calculada.
	// Nota: Si en el listado NO necesitas el detalle por almacén,
	// puedes quitar la columna 'inventory_json' de aquí para hacerlo aún más ligero.
	query := `
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
             COALESCE(a.image_url, ''),
             COALESCE(a.inventory_json, '{}') -- ¡Lectura directa!
        FROM art a
        WHERE (a.art_des ILIKE $1 OR a.co_art ILIKE $1 OR a.campo4 ILIKE $1)
        ORDER BY a.art_des ASC
        LIMIT $2 OFFSET $3
     `

	searchTerm := "%" + search + "%"
	rows, err := r.db.Query(query, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error en el query: %w", err)
	}
	defer rows.Close()

	products := []Product{}

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.CoArt, &p.ArtDes, &p.StockAct,
			&p.PrecVta1, &p.PrecVta2, &p.PrecVta3, &p.PrecVta4, &p.PrecVta5,
			&p.TipoImp, &p.CoLin, &p.CoCat, &p.CoSubl,
			&p.ImageUrl, &p.InventarioRaw,
		); err != nil {
			return nil, err
		}

		if len(p.InventarioRaw) > 0 {
			_ = json.Unmarshal(p.InventarioRaw, &p.Inventario)
		}

		products = append(products, p)
	}

	return products, nil
}
