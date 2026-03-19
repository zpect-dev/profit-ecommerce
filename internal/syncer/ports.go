package syncer

import "context"

// pageSize define el tamaño de página para lecturas masivas paginadas.
const pageSize = 2000

// SourceRepo define el contrato de lectura contra la BD origen (Profit/SQL Server).
// El Service depende de esta interfaz, no de la implementación concreta.
type SourceRepo interface {
	FetchLinArt(ctx context.Context) ([]LinArt, error)
	FetchCatArt(ctx context.Context) ([]CatArt, error)
	FetchSubLin(ctx context.Context) ([]SubLin, error)
	FetchAlmacen(ctx context.Context) ([]Almacen, error)
	FetchSubAlma(ctx context.Context) ([]SubAlma, error)
	FetchDescuentos(ctx context.Context) ([]Descuento, error)
	FetchArticlesPage(ctx context.Context, limit, offset int) ([]Article, error)
	FetchStAlmacPage(ctx context.Context, limit, offset int) ([]StAlmac, error)
}

// DestRepo define el contrato de escritura contra la BD destino (PostgreSQL).
// El Service depende de esta interfaz, no de la implementación concreta.
type DestRepo interface {
	UpsertLinArt(ctx context.Context, items []LinArt) (int, error)
	UpsertCatArt(ctx context.Context, items []CatArt) (int, error)
	UpsertSubLin(ctx context.Context, items []SubLin) (int, error)
	UpsertAlmacen(ctx context.Context, items []Almacen) (int, error)
	UpsertSubAlma(ctx context.Context, items []SubAlma) (int, error)
	TruncateAndInsertDescuentos(ctx context.Context, items []Descuento) (int, error)
	UpsertArticles(ctx context.Context, items []Article) (int, error)
	UpsertStAlmac(ctx context.Context, items []StAlmac) (int, error)
	RecalculateInventoryJSON(ctx context.Context) error
}
