package syncer

import (
	"context"
	"fmt"
	"log"
)

// Service orquesta la sincronización entre Profit (origen) y PostgreSQL (destino).
// Depende de interfaces (SourceRepo, DestRepo), no de implementaciones concretas.
type Service struct {
	source SourceRepo
	dest   DestRepo
}

// NewService crea un nuevo servicio de sincronización inyectando las interfaces de repositorio.
func NewService(source SourceRepo, dest DestRepo) *Service {
	return &Service{source: source, dest: dest}
}

// GenerateImageURL genera la URL de imagen a partir de un código de artículo.
// Lógica de negocio: quita el sufijo 'C' o 'A' del código y construye la URL.
func GenerateImageURL(coArt string) string {
	imageCo := coArt
	if len(coArt) > 0 && (coArt[len(coArt)-1] == 'C' || coArt[len(coArt)-1] == 'A') {
		imageCo = coArt[:len(coArt)-1]
	}
	return fmt.Sprintf("https://imagenes.cristmedicals.com/imagenes-v3/imagenes/%s.jpg", imageCo)
}

// RunSlowSync sincroniza las tablas "estáticas" o de configuración (maestros).
// Se ejecuta con poca frecuencia (ej. cada 1 hora).
// Respeta la cancelación del contexto: si ctx se cancela, el ciclo se detiene.
func (s *Service) RunSlowSync(ctx context.Context) {
	fmt.Println("--- Slow Sync: Maestros ---")

	// 1. Líneas
	if err := ctx.Err(); err != nil {
		log.Printf("Slow Sync cancelado antes de lin_art: %v", err)
		return
	}
	if items, err := s.source.FetchLinArt(ctx); err != nil {
		log.Printf("Error fetching lin_art: %v", err)
	} else {
		count, _ := s.dest.UpsertLinArt(ctx, items)
		fmt.Printf("Lineas sincronizadas: %d\n", count)
	}

	// 2. Categorías
	if err := ctx.Err(); err != nil {
		log.Printf("Slow Sync cancelado antes de cat_art: %v", err)
		return
	}
	if items, err := s.source.FetchCatArt(ctx); err != nil {
		log.Printf("Error fetching cat_art: %v", err)
	} else {
		count, _ := s.dest.UpsertCatArt(ctx, items)
		fmt.Printf("Categorias sincronizadas: %d\n", count)
	}

	// 3. Sub-líneas
	if err := ctx.Err(); err != nil {
		log.Printf("Slow Sync cancelado antes de sub_lin: %v", err)
		return
	}
	if items, err := s.source.FetchSubLin(ctx); err != nil {
		log.Printf("Error fetching sub_lin: %v", err)
	} else {
		count, _ := s.dest.UpsertSubLin(ctx, items)
		fmt.Printf("Sub-lineas sincronizadas: %d\n", count)
	}

	// 4. Almacenes
	if err := ctx.Err(); err != nil {
		log.Printf("Slow Sync cancelado antes de almacen: %v", err)
		return
	}
	if items, err := s.source.FetchAlmacen(ctx); err != nil {
		log.Printf("Error fetching almacen: %v", err)
	} else {
		count, _ := s.dest.UpsertAlmacen(ctx, items)
		fmt.Printf("Almacenes sincronizados: %d\n", count)
	}

	// 5. Sub-almacenes
	if err := ctx.Err(); err != nil {
		log.Printf("Slow Sync cancelado antes de sub_alma: %v", err)
		return
	}
	if items, err := s.source.FetchSubAlma(ctx); err != nil {
		log.Printf("Error fetching sub_alma: %v", err)
	} else {
		count, _ := s.dest.UpsertSubAlma(ctx, items)
		fmt.Printf("Sub-almacenes sincronizados: %d\n", count)
	}

	fmt.Println("--- Slow Sync: Completado ---")
}

// RunFastSync sincroniza las tablas críticas de venta (artículos, stock, precios, descuentos).
// Se ejecuta con alta frecuencia (ej. cada 1 minuto).
// Usa paginación (pageSize) para no cargar tablas grandes en memoria de golpe.
func (s *Service) RunFastSync(ctx context.Context) {
	fmt.Println("--- Fast Sync: Stock/Precios ---")

	// 1. Descuentos
	if err := ctx.Err(); err != nil {
		log.Printf("Fast Sync cancelado antes de descuentos: %v", err)
		return
	}
	if items, err := s.source.FetchDescuentos(ctx); err != nil {
		log.Printf("Error fetching descuentos: %v", err)
	} else {
		count, _ := s.dest.TruncateAndInsertDescuentos(ctx, items)
		fmt.Printf("Descuentos sincronizados: %d\n", count)
	}

	// 2. Artículos (paginado)
	if err := ctx.Err(); err != nil {
		log.Printf("Fast Sync cancelado antes de articulos: %v", err)
		return
	}
	s.syncArticlesPaginated(ctx)

	// 3. Stock por almacén (paginado)
	if err := ctx.Err(); err != nil {
		log.Printf("Fast Sync cancelado antes de st_almac: %v", err)
		return
	}
	s.syncStAlmacPaginated(ctx)

	// 4. Recalcular JSON de inventario
	if err := ctx.Err(); err != nil {
		log.Printf("Fast Sync cancelado antes de recalcular JSON: %v", err)
		return
	}
	fmt.Print("Recalculando JSON de Inventario... ")
	if err := s.dest.RecalculateInventoryJSON(ctx); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("OK")
	}

	fmt.Println("--- Fast Sync: Completado ---")
}

// syncArticlesPaginated lee artículos en páginas de pageSize, aplica la transformación
// de ImageURL (lógica de negocio), y los envía al repositorio destino.
func (s *Service) syncArticlesPaginated(ctx context.Context) {
	totalCount := 0
	offset := 0

	for {
		if err := ctx.Err(); err != nil {
			log.Printf("Sync artículos cancelado en offset %d: %v", offset, err)
			return
		}

		page, err := s.source.FetchArticlesPage(ctx, pageSize, offset)
		if err != nil {
			log.Printf("Error fetching articles (offset=%d): %v", offset, err)
			return
		}

		if len(page) == 0 {
			break // No hay más páginas
		}

		// Lógica de negocio: generar ImageURL para cada artículo
		for i := range page {
			page[i].ImageURL = GenerateImageURL(page[i].CoArt)
		}

		count, _ := s.dest.UpsertArticles(ctx, page)
		totalCount += count
		offset += len(page)

		fmt.Printf("  Artículos sincronizados: página offset=%d, filas=%d\n", offset-len(page), count)

		if len(page) < pageSize {
			break // Última página parcial
		}
	}

	fmt.Printf("Articulos sincronizados total: %d\n", totalCount)
}

// syncStAlmacPaginated lee stock por almacén en páginas de pageSize.
func (s *Service) syncStAlmacPaginated(ctx context.Context) {
	totalCount := 0
	offset := 0

	for {
		if err := ctx.Err(); err != nil {
			log.Printf("Sync st_almac cancelado en offset %d: %v", offset, err)
			return
		}

		page, err := s.source.FetchStAlmacPage(ctx, pageSize, offset)
		if err != nil {
			log.Printf("Error fetching st_almac (offset=%d): %v", offset, err)
			return
		}

		if len(page) == 0 {
			break
		}

		count, _ := s.dest.UpsertStAlmac(ctx, page)
		totalCount += count
		offset += len(page)

		fmt.Printf("  Stock almacen sincronizado: página offset=%d, filas=%d\n", offset-len(page), count)

		if len(page) < pageSize {
			break
		}
	}

	fmt.Printf("Stock almacen sincronizado total: %d\n", totalCount)
}
