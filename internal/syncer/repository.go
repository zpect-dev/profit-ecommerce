package syncer

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// SourceRepository lee datos del sistema Profit (SQL Server).
type SourceRepository struct {
	db *sqlx.DB
}

// NewSourceRepository crea un nuevo repositorio de lectura contra la BD origen.
func NewSourceRepository(db *sqlx.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

// DestRepository escribe datos en PostgreSQL (BD destino).
type DestRepository struct {
	db *sqlx.DB
}

// NewDestRepository crea un nuevo repositorio de escritura contra PostgreSQL.
func NewDestRepository(db *sqlx.DB) *DestRepository {
	return &DestRepository{db: db}
}

// --- SOURCE: Lectura desde Profit ---

func (r *SourceRepository) FetchLinArt(ctx context.Context) ([]LinArt, error) {
	var items []LinArt
	rows, err := r.db.QueryContext(ctx, "SELECT co_lin, lin_des FROM lin_art")
	if err != nil {
		return nil, fmt.Errorf("error fetching lin_art: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item LinArt
		if err := rows.Scan(&item.CoLin, &item.LinDes); err != nil {
			log.Printf("Error scan lin_art: %v", err)
			continue
		}
		item.CoLin = strings.TrimSpace(item.CoLin)
		item.LinDes = strings.TrimSpace(item.LinDes)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchCatArt(ctx context.Context) ([]CatArt, error) {
	var items []CatArt
	rows, err := r.db.QueryContext(ctx, "SELECT co_cat, cat_des FROM cat_art")
	if err != nil {
		return nil, fmt.Errorf("error fetching cat_art: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item CatArt
		if err := rows.Scan(&item.CoCat, &item.CatDes); err != nil {
			log.Printf("Error scan cat_art: %v", err)
			continue
		}
		item.CoCat = strings.TrimSpace(item.CoCat)
		item.CatDes = strings.TrimSpace(item.CatDes)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchSubLin(ctx context.Context) ([]SubLin, error) {
	var items []SubLin
	rows, err := r.db.QueryContext(ctx, "SELECT co_subl, subl_des, co_lin FROM sub_lin")
	if err != nil {
		return nil, fmt.Errorf("error fetching sub_lin: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item SubLin
		if err := rows.Scan(&item.CoSubl, &item.SublDes, &item.CoLin); err != nil {
			log.Printf("Error scan sub_lin: %v", err)
			continue
		}
		item.CoSubl = strings.TrimSpace(item.CoSubl)
		item.SublDes = strings.TrimSpace(item.SublDes)
		item.CoLin = strings.TrimSpace(item.CoLin)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchAlmacen(ctx context.Context) ([]Almacen, error) {
	var items []Almacen
	rows, err := r.db.QueryContext(ctx, "SELECT co_alma, alma_des FROM almacen")
	if err != nil {
		return nil, fmt.Errorf("error fetching almacen: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Almacen
		if err := rows.Scan(&item.CoAlma, &item.AlmaDes); err != nil {
			log.Printf("Error scan almacen: %v", err)
			continue
		}
		item.CoAlma = strings.TrimSpace(item.CoAlma)
		item.AlmaDes = strings.TrimSpace(item.AlmaDes)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchSubAlma(ctx context.Context) ([]SubAlma, error) {
	var items []SubAlma
	rows, err := r.db.QueryContext(ctx, "SELECT co_sub, des_sub, co_alma FROM sub_alma")
	if err != nil {
		return nil, fmt.Errorf("error fetching sub_alma: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item SubAlma
		if err := rows.Scan(&item.CoSub, &item.DesSub, &item.CoAlma); err != nil {
			log.Printf("Error scan sub_alma: %v", err)
			continue
		}
		item.CoSub = strings.TrimSpace(item.CoSub)
		item.DesSub = strings.TrimSpace(item.DesSub)
		item.CoAlma = strings.TrimSpace(item.CoAlma)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchDescuentos(ctx context.Context) ([]Descuento, error) {
	var items []Descuento
	rows, err := r.db.QueryContext(ctx, "SELECT co_desc, tipo_cli, tipo_desc, porc1 FROM descuen")
	if err != nil {
		return nil, fmt.Errorf("error fetching descuen: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Descuento
		if err := rows.Scan(&item.CoDesc, &item.TipoCli, &item.TipoDesc, &item.Porc1); err != nil {
			log.Printf("Error scan descuen: %v", err)
			continue
		}
		item.CoDesc = strings.TrimSpace(item.CoDesc)
		item.TipoCli = strings.TrimSpace(item.TipoCli)
		item.TipoDesc = strings.TrimSpace(item.TipoDesc)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchArticlesPage(ctx context.Context, limit, offset int) ([]Article, error) {
	query := `
		SELECT
			co_art, 
			art_des,
			stock_act,
			prec_vta1,
			prec_vta2,
			prec_vta3,
			prec_vta4,
			prec_vta5,
			tipo_imp,
			COALESCE(co_lin, ''),
			COALESCE(co_cat, ''),
			COALESCE(co_subl, ''),
			COALESCE(campo4, '')
		FROM art
		WHERE anulado = 0 AND art_des NOT LIKE '%NO USAR%'
		ORDER BY co_art
		OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY
	`
	rows, err := r.db.QueryContext(ctx, query, sql.Named("offset", offset), sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("error fetching articles page (offset=%d): %w", offset, err)
	}
	defer rows.Close()

	var items []Article
	for rows.Next() {
		var item Article
		if err := rows.Scan(
			&item.CoArt, &item.ArtDes, &item.StockAct,
			&item.PrecVta1, &item.PrecVta2, &item.PrecVta3, &item.PrecVta4, &item.PrecVta5,
			&item.TipoImp, &item.CoLin, &item.CoCat, &item.CoSubl, &item.Campo4,
		); err != nil {
			log.Printf("Error scan article: %v", err)
			continue
		}
		item.CoArt = strings.TrimSpace(item.CoArt)
		item.ArtDes = strings.TrimSpace(item.ArtDes)
		item.TipoImp = strings.TrimSpace(item.TipoImp)
		item.CoLin = strings.TrimSpace(item.CoLin)
		item.CoCat = strings.TrimSpace(item.CoCat)
		item.CoSubl = strings.TrimSpace(item.CoSubl)
		item.Campo4 = strings.TrimSpace(item.Campo4)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchStAlmacPage(ctx context.Context, limit, offset int) ([]StAlmac, error) {
	query := `
		SELECT 
			s.co_alma, s.co_art, 
			s.stock_act, s.sstock_act, 
			s.stock_com, s.sstock_com, 
			s.stock_lle, s.sstock_lle, 
			s.stock_des, s.sstock_des 
		FROM st_almac s
		INNER JOIN art a ON s.co_art = a.co_art
		WHERE a.anulado = 0 AND a.art_des NOT LIKE '%NO USAR%'
		ORDER BY s.co_art, s.co_alma
		OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY
	`
	rows, err := r.db.QueryContext(ctx, query, sql.Named("offset", offset), sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("error fetching st_almac page (offset=%d): %w", offset, err)
	}
	defer rows.Close()

	var items []StAlmac
	for rows.Next() {
		var item StAlmac
		if err := rows.Scan(
			&item.CoAlma, &item.CoArt,
			&item.StockAct, &item.SStockAct,
			&item.StockCom, &item.SStockCom,
			&item.StockLle, &item.SStockLle,
			&item.StockDes, &item.SStockDes,
		); err != nil {
			log.Printf("Error scan st_almac: %v", err)
			continue
		}
		item.CoAlma = strings.TrimSpace(item.CoAlma)
		item.CoArt = strings.TrimSpace(item.CoArt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchTiposCli(ctx context.Context) ([]TipoCli, error) {
	var items []TipoCli
	rows, err := r.db.QueryContext(ctx, "SELECT tip_cli, des_tipo, precio_a FROM tipo_cli")
	if err != nil {
		return nil, fmt.Errorf("error fetching tipo_cli: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item TipoCli
		if err := rows.Scan(&item.TipCli, &item.DesTipo, &item.PrecioA); err != nil {
			log.Printf("Error scan tipo_cli: %v", err)
			continue
		}
		item.TipCli = strings.TrimSpace(item.TipCli)
		item.DesTipo = strings.TrimSpace(item.DesTipo)
		item.PrecioA = strings.TrimSpace(item.PrecioA)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceRepository) FetchClientesPage(ctx context.Context, limit, offset int) ([]Cliente, error) {
	query := `
		SELECT co_cli, tipo, cli_des, rif, inactivo, login, mont_cre 
		FROM clientes 
		ORDER BY co_cli 
		OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY
	`
	rows, err := r.db.QueryContext(ctx, query, sql.Named("offset", offset), sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("error fetching clientes page (offset=%d): %w", offset, err)
	}
	defer rows.Close()

	var items []Cliente
	for rows.Next() {
		var item Cliente
		var loginStr, montCreStr sql.NullString
		
		if err := rows.Scan(&item.CoCli, &item.Tipo, &item.CliDes, &item.Rif, &item.Inactivo, &loginStr, &montCreStr); err != nil {
			log.Printf("Error scan cliente: %v", err)
			continue
		}
		item.CoCli = strings.TrimSpace(item.CoCli)
		item.Tipo = strings.TrimSpace(item.Tipo)
		item.CliDes = strings.TrimSpace(item.CliDes)
		item.Rif = strings.TrimSpace(item.Rif)

		if loginStr.Valid {
			if val, err := strconv.ParseFloat(strings.TrimSpace(loginStr.String), 64); err == nil {
				item.Login = val
			}
		}
		if montCreStr.Valid {
			if val, err := strconv.ParseFloat(strings.TrimSpace(montCreStr.String), 64); err == nil {
				item.MontCre = val
			}
		}

		items = append(items, item)
	}
	return items, rows.Err()
}

// --- DEST: Escritura en PostgreSQL (Batch Inserts) ---

// batchSize define cuántas filas se insertan por query batch.
// PostgreSQL soporta hasta 65535 parámetros por query.
// Con la tabla más ancha (art, 14 columnas): 65535/14 = 4681 filas max. 500 es seguro.
const batchSize = 500

// buildPlaceholders genera los placeholders ($1,$2),($3,$4)... para N filas con colsPerRow columnas.
func buildPlaceholders(numRows, colsPerRow int) string {
	rows := make([]string, 0, numRows)
	paramIdx := 1
	for i := 0; i < numRows; i++ {
		cols := make([]string, 0, colsPerRow)
		for j := 0; j < colsPerRow; j++ {
			cols = append(cols, fmt.Sprintf("$%d", paramIdx))
			paramIdx++
		}
		rows = append(rows, "("+strings.Join(cols, ",")+")")
	}
	return strings.Join(rows, ",")
}

// execBatchWithFallback intenta el batch INSERT. Si falla (ej. FK violation en 1 fila),
// hace fallback a fila-por-fila para ese chunk, logueando las filas individuales que fallen.
// queryTemplate debe tener un %s donde van los placeholders VALUES.
func (r *DestRepository) execBatchWithFallback(ctx context.Context, queryTemplate string, args []interface{}, colsPerRow int) int {
	// Intento 1: batch completo
	numRows := len(args) / colsPerRow
	batchPlaceholders := buildPlaceholders(numRows, colsPerRow)
	batchQuery := fmt.Sprintf(queryTemplate, batchPlaceholders)

	_, err := r.db.ExecContext(ctx, batchQuery, args...)
	if err == nil {
		return numRows
	}

	// Intento 2: fallback fila por fila para este chunk
	log.Printf("Batch falló, reintentando %d filas individualmente... (%v)", numRows, err)
	count := 0
	for row := 0; row < numRows; row++ {
		start := row * colsPerRow
		end := start + colsPerRow
		rowArgs := args[start:end]

		singlePlaceholders := buildPlaceholders(1, colsPerRow)
		query := fmt.Sprintf(queryTemplate, singlePlaceholders)

		_, err := r.db.ExecContext(ctx, query, rowArgs...)
		if err != nil {
			log.Printf("  Fila ignorada: %v", err)
		} else {
			count++
		}
	}
	return count
}

func (r *DestRepository) UpsertLinArt(ctx context.Context, items []LinArt) (int, error) {
	const cols = 2
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]


		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.CoLin, item.LinDes)
		}

		queryTpl := `
			INSERT INTO lin_art (co_lin, lin_des) VALUES %s
			ON CONFLICT (co_lin) DO UPDATE SET lin_des = EXCLUDED.lin_des
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

func (r *DestRepository) UpsertCatArt(ctx context.Context, items []CatArt) (int, error) {
	const cols = 2
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.CoCat, item.CatDes)
		}

		queryTpl := `
			INSERT INTO cat_art (co_cat, cat_des) VALUES %s
			ON CONFLICT (co_cat) DO UPDATE SET cat_des = EXCLUDED.cat_des
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

func (r *DestRepository) UpsertSubLin(ctx context.Context, items []SubLin) (int, error) {
	const cols = 3
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.CoSubl, item.SublDes, item.CoLin)
		}

		queryTpl := `
			INSERT INTO sub_lin (co_subl, subl_des, co_lin) VALUES %s
			ON CONFLICT (co_subl) DO UPDATE SET subl_des = EXCLUDED.subl_des, co_lin = EXCLUDED.co_lin
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

func (r *DestRepository) UpsertAlmacen(ctx context.Context, items []Almacen) (int, error) {
	const cols = 2
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.CoAlma, item.AlmaDes)
		}

		queryTpl := `
			INSERT INTO almacen (co_alma, alma_des) VALUES %s
			ON CONFLICT (co_alma) DO UPDATE SET alma_des = EXCLUDED.alma_des
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

func (r *DestRepository) UpsertSubAlma(ctx context.Context, items []SubAlma) (int, error) {
	const cols = 3
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.CoSub, item.DesSub, item.CoAlma)
		}

		queryTpl := `
			INSERT INTO sub_alma (co_sub, des_sub, co_alma) VALUES %s
			ON CONFLICT (co_sub) DO UPDATE SET des_sub = EXCLUDED.des_sub, co_alma = EXCLUDED.co_alma
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

// TruncateAndInsertDescuentos ejecuta TRUNCATE + INSERT dentro de una transacción explícita.
// Si cualquier batch falla, se hace Rollback y la tabla queda intacta.
func (r *DestRepository) TruncateAndInsertDescuentos(ctx context.Context, items []Descuento) (int, error) {
	const cols = 4

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("error iniciando transacción descuentos: %w", err)
	}
	// Rollback seguro: si ya se hizo Commit, Rollback es un no-op
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE descuen")
	if err != nil {
		return 0, fmt.Errorf("error truncando descuen: %w", err)
	}

	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		placeholders := buildPlaceholders(len(chunk), cols)
		query := fmt.Sprintf(`
			INSERT INTO descuen (co_desc, tipo_cli, tipo_desc, porc1) VALUES %s
		`, placeholders)

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.CoDesc, item.TipoCli, item.TipoDesc, item.Porc1)
		}

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return count, fmt.Errorf("error batch descuen (filas %d-%d): %w", i, end-1, err)
		}
		count += len(chunk)
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("error commit descuentos: %w", err)
	}

	return count, nil
}

func (r *DestRepository) UpsertArticles(ctx context.Context, items []Article) (int, error) {
	const cols = 14
	count := 0

	toNull := func(s string) sql.NullString {
		if s == "" {
			return sql.NullString{}
		}
		return sql.NullString{String: s, Valid: true}
	}

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args,
				item.CoArt, item.ArtDes, item.StockAct,
				item.PrecVta1, item.PrecVta2, item.PrecVta3, item.PrecVta4, item.PrecVta5,
				item.TipoImp,
				toNull(item.CoLin), toNull(item.CoCat), toNull(item.CoSubl), toNull(item.Campo4),
				item.ImageURL,
			)
		}

		queryTpl := `
			INSERT INTO art (co_art, art_des, stock_act, prec_vta1, prec_vta2, prec_vta3, prec_vta4, prec_vta5, tipo_imp, co_lin, co_cat, co_subl, campo4, image_url)
			VALUES %s
			ON CONFLICT (co_art) DO UPDATE SET
				art_des = EXCLUDED.art_des,
				stock_act = EXCLUDED.stock_act,
				prec_vta1 = EXCLUDED.prec_vta1,
				prec_vta2 = EXCLUDED.prec_vta2,
				prec_vta3 = EXCLUDED.prec_vta3,
				prec_vta4 = EXCLUDED.prec_vta4,
				prec_vta5 = EXCLUDED.prec_vta5,
				tipo_imp = EXCLUDED.tipo_imp,
				co_lin = EXCLUDED.co_lin,
				co_cat = EXCLUDED.co_cat,
				co_subl = EXCLUDED.co_subl,
				campo4 = EXCLUDED.campo4,
				image_url = EXCLUDED.image_url,
				last_sync = NOW()
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

func (r *DestRepository) UpsertStAlmac(ctx context.Context, items []StAlmac) (int, error) {
	const cols = 10
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args,
				item.CoAlma, item.CoArt,
				item.StockAct, item.SStockAct,
				item.StockCom, item.SStockCom,
				item.StockLle, item.SStockLle,
				item.StockDes, item.SStockDes,
			)
		}

		queryTpl := `
			INSERT INTO st_almac (co_alma, co_art, stock_act, sstock_act, stock_com, sstock_com, stock_lle, sstock_lle, stock_des, sstock_des) 
			VALUES %s
			ON CONFLICT (co_alma, co_art) DO UPDATE SET
				stock_act = EXCLUDED.stock_act,
				sstock_act = EXCLUDED.sstock_act,
				stock_com = EXCLUDED.stock_com,
				sstock_com = EXCLUDED.sstock_com,
				stock_lle = EXCLUDED.stock_lle,
				sstock_lle = EXCLUDED.sstock_lle,
				stock_des = EXCLUDED.stock_des,
				sstock_des = EXCLUDED.sstock_des
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

// RecalculateInventoryJSON recalcula el JSONB de inventario agregado en la tabla art.
func (r *DestRepository) RecalculateInventoryJSON(ctx context.Context) error {
	queryJSON := `
		UPDATE art p
		SET inventory_json = subquery.json_data
		FROM (
			SELECT 
				pre_calculated.co_art,
				jsonb_object_agg(
					TRIM(pre_calculated.co_alma),
					jsonb_build_object(
						'nombre', TRIM(pre_calculated.alma_des),
						'stock_total', pre_calculated.total_act,
						'stock_comprometido', pre_calculated.total_com,
						'stock_por_llegar', pre_calculated.total_lle
					)
				) as json_data
			FROM (
				SELECT 
					st.co_art,
					a.co_alma,
					a.alma_des,
					SUM(st.stock_act) as total_act,
					SUM(st.stock_com) as total_com,
					SUM(st.stock_lle) as total_lle
				FROM st_almac st
				JOIN sub_alma sa ON st.co_alma = sa.co_sub
				JOIN almacen a ON sa.co_alma = a.co_alma
				GROUP BY st.co_art, a.co_alma, a.alma_des
				HAVING SUM(st.stock_act) > 0
			) pre_calculated
			GROUP BY pre_calculated.co_art
		) AS subquery
		WHERE p.co_art = subquery.co_art;
	`
	_, err := r.db.ExecContext(ctx, queryJSON)
	if err != nil {
		return fmt.Errorf("error actualizando inventory JSON: %w", err)
	}
	return nil
}

func (r *DestRepository) UpsertTiposCli(ctx context.Context, items []TipoCli) (int, error) {
	const cols = 3
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.TipCli, item.DesTipo, item.PrecioA)
		}

		queryTpl := `
			INSERT INTO tipo_cli (tip_cli, des_tipo, precio_a) VALUES %s
			ON CONFLICT (tip_cli) DO UPDATE SET 
				des_tipo = EXCLUDED.des_tipo, 
				precio_a = EXCLUDED.precio_a
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}

func (r *DestRepository) UpsertClientes(ctx context.Context, items []Cliente) (int, error) {
	const cols = 7
	count := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]

		args := make([]interface{}, 0, len(chunk)*cols)
		for _, item := range chunk {
			args = append(args, item.CoCli, item.Tipo, item.CliDes, item.Rif, item.Inactivo, item.Login, item.MontCre)
		}

		queryTpl := `
			INSERT INTO clientes (co_cli, tipo, cli_des, rif, inactivo, login, mont_cre) VALUES %s
			ON CONFLICT (co_cli) DO UPDATE SET 
				tipo = EXCLUDED.tipo, 
				cli_des = EXCLUDED.cli_des, 
				rif = EXCLUDED.rif, 
				inactivo = EXCLUDED.inactivo, 
				login = EXCLUDED.login,
				mont_cre = EXCLUDED.mont_cre
		`
		count += r.execBatchWithFallback(ctx, queryTpl, args, cols)
	}
	return count, nil
}
