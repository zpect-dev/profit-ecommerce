package catalog

import "context"

// CatalogRepository define el contrato para acceder al catálogo de productos.
// Los handlers y services dependen de esta interfaz, no de la implementación concreta.
type CatalogRepository interface {
	ListProducts(ctx context.Context, page, limit int, search, category string, inStock, hasDiscount bool) ([]Product, error)
	GetByID(ctx context.Context, id string) (Product, error)
	GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error)
	ListCategories(ctx context.Context) ([]Category, error)
}
