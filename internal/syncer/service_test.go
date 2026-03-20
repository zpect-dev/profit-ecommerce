package syncer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

// --- MOCKS ---

type mockSourceRepo struct {
	mock.Mock
}

func (m *mockSourceRepo) FetchLinArt(ctx context.Context) ([]LinArt, error)          { return nil, nil }
func (m *mockSourceRepo) FetchCatArt(ctx context.Context) ([]CatArt, error)          { return nil, nil }
func (m *mockSourceRepo) FetchSubLin(ctx context.Context) ([]SubLin, error)          { return nil, nil }
func (m *mockSourceRepo) FetchAlmacen(ctx context.Context) ([]Almacen, error)        { return nil, nil }
func (m *mockSourceRepo) FetchSubAlma(ctx context.Context) ([]SubAlma, error)        { return nil, nil }
func (m *mockSourceRepo) FetchDescuentos(ctx context.Context) ([]Descuento, error)   { return nil, nil }
func (m *mockSourceRepo) FetchTiposCli(ctx context.Context) ([]TipoCli, error)       { return nil, nil }
func (m *mockSourceRepo) FetchStAlmacPage(ctx context.Context, limit, offset int) ([]StAlmac, error) {
	return nil, nil
}
func (m *mockSourceRepo) FetchClientesPage(ctx context.Context, limit, offset int) ([]Cliente, error) {
	return nil, nil
}

func (m *mockSourceRepo) FetchArticlesPage(ctx context.Context, limit, offset int) ([]Article, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) != nil {
		return args.Get(0).([]Article), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockDestRepo struct {
	mock.Mock
}

func (m *mockDestRepo) UpsertLinArt(ctx context.Context, items []LinArt) (int, error)                 { return 0, nil }
func (m *mockDestRepo) UpsertCatArt(ctx context.Context, items []CatArt) (int, error)                 { return 0, nil }
func (m *mockDestRepo) UpsertSubLin(ctx context.Context, items []SubLin) (int, error)                 { return 0, nil }
func (m *mockDestRepo) UpsertAlmacen(ctx context.Context, items []Almacen) (int, error)               { return 0, nil }
func (m *mockDestRepo) UpsertSubAlma(ctx context.Context, items []SubAlma) (int, error)               { return 0, nil }
func (m *mockDestRepo) TruncateAndInsertDescuentos(ctx context.Context, items []Descuento) (int, error) {
	return 0, nil
}
func (m *mockDestRepo) UpsertTiposCli(ctx context.Context, items []TipoCli) (int, error)                 { return 0, nil }
func (m *mockDestRepo) UpsertStAlmac(ctx context.Context, items []StAlmac) (int, error) { return 0, nil }
func (m *mockDestRepo) UpsertClientes(ctx context.Context, items []Cliente) (int, error) { return 0, nil }
func (m *mockDestRepo) RecalculateInventoryJSON(ctx context.Context) error              { return nil }

func (m *mockDestRepo) UpsertArticles(ctx context.Context, items []Article) (int, error) {
	args := m.Called(ctx, items)
	return args.Int(0), args.Error(1)
}

// --- TESTS: Pagination y Sincronización ---

func TestService_syncArticlesPaginated(t *testing.T) {
	t.Parallel()

	// Creamos un dummy payload de tamaño exacto 'pageSize' para forzar una página completa
	fullPage := make([]Article, pageSize)
	for i := 0; i < pageSize; i++ {
		fullPage[i] = Article{CoArt: "ART-TEST"} // Setup básico
	}
	// Payload de media página para salir del loop
	halfPage := make([]Article, pageSize/2)
	for i := 0; i < pageSize/2; i++ {
		halfPage[i] = Article{CoArt: "ART-HALF"}
	}

	tests := []struct {
		name      string
		setupMock func(src *mockSourceRepo, dst *mockDestRepo, ctx context.Context)
		cancelCtx bool // Activa cancelación intencionada de contexto
	}{
		{
			name: "Éxito Completo: El origen devuelve 2 páginas",
			setupMock: func(src *mockSourceRepo, dst *mockDestRepo, ctx context.Context) {
				// Página 1 (completa)
				src.On("FetchArticlesPage", mock.Anything, pageSize, 0).Return(fullPage, nil).Once()
				dst.On("UpsertArticles", mock.Anything, mock.MatchedBy(func(items []Article) bool {
					return len(items) == pageSize && items[0].ImageURL == GenerateImageURL("ART-TEST") // Validar lógica de negocio de imagen
				})).Return(pageSize, nil).Once()

				// Página 2 (incompleta, detiene el loop)
				src.On("FetchArticlesPage", mock.Anything, pageSize, pageSize).Return(halfPage, nil).Once()
				dst.On("UpsertArticles", mock.Anything, mock.MatchedBy(func(items []Article) bool {
					return len(items) == pageSize/2
				})).Return(pageSize/2, nil).Once()
			},
			cancelCtx: false,
		},
		{
			name: "Error de Origen: Fallo capa inferior de DB",
			setupMock: func(src *mockSourceRepo, dst *mockDestRepo, ctx context.Context) {
				// Simular fallo de conexión al origen
				src.On("FetchArticlesPage", mock.Anything, pageSize, 0).Return(nil, errors.New("timeout SQL Server origen")).Once()
				// Dest nunca debería ser llamado
				dst.AssertNotCalled(t, "UpsertArticles")
			},
			cancelCtx: false,
		},
		{
			name: "Error de Destino (Batch Insert): Constraint aborts saving",
			setupMock: func(src *mockSourceRepo, dst *mockDestRepo, ctx context.Context) {
				src.On("FetchArticlesPage", mock.Anything, pageSize, 0).Return(halfPage, nil).Once()
				// Destino falla al insertar. Como no hace panic ni hace return explícito en `service.go`,
				// comprobamos que el logging maneje el _ silenciosamente y el proceso no se caiga.
				dst.On("UpsertArticles", mock.Anything, mock.Anything).Return(0, errors.New("psql: constraint violation")).Once()
			},
			cancelCtx: false,
		},
		{
			name: "Caso Borde (Context Cancellation): Cancelación en mitad de paginación",
			setupMock: func(src *mockSourceRepo, dst *mockDestRepo, ctx context.Context) {
				// El test correrá cancelando el contexto en test execution (fuera del mock). 
				// El ciclo `for` en `syncArticlesPaginated` debe respetar `ctx.Err() != nil`
				// y salir INMEDIATAMENTE sin llegar siquiera a `FetchArticlesPage` inicial.
				
				// Como el contexto entra al test ya cancelado (cancelCtx: true), nunca llamará a Fetch ni a Upsert.
				src.AssertNotCalled(t, "FetchArticlesPage")
				dst.AssertNotCalled(t, "UpsertArticles")
			},
			cancelCtx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := new(mockSourceRepo)
			dst := new(mockDestRepo)

			var ctx context.Context
			var cancel context.CancelFunc

			if tt.cancelCtx {
				ctx, cancel = context.WithCancel(context.Background())
				cancel() // Se cancela el contexto ANTES de pasarlo, forzando la muerte súbita
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			tt.setupMock(src, dst, ctx)

			s := NewService(src, dst)
			s.syncArticlesPaginated(ctx)

			src.AssertExpectations(t)
			dst.AssertExpectations(t)
		})
	}
}
