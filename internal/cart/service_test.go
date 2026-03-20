package cart

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCKS ---

type mockCacheRepo struct {
	mock.Mock
}

func (m *mockCacheRepo) SaveCart(ctx context.Context, cart Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *mockCacheRepo) GetCart(ctx context.Context, userID string) (Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(Cart), args.Error(1)
}

func (m *mockCacheRepo) DeleteCart(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockDBRepo struct {
	mock.Mock
}

func (m *mockDBRepo) PersistCart(ctx context.Context, cart Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

type mockCatalog struct {
	mock.Mock
}

func (m *mockCatalog) CheckStock(ctx context.Context, productIDs []string) (map[string]int, error) {
	args := m.Called(ctx, productIDs)
	return args.Get(0).(map[string]int), args.Error(1)
}

// --- TESTS: AddToCart ---

func TestService_AddToCart(t *testing.T) {
	userID := "user-bdd"

	tests := []struct {
		name          string
		initialCart   Cart
		initialErr    error
		itemToAdd     CartItem
		mockSaveErr   error
		simulateBlock bool
		expectedError string
		validateCart  func(t *testing.T, cart Cart)
	}{
		{
			name:        "Éxito: El carrito no existe (redis: nil)",
			initialCart: Cart{},
			initialErr:  errors.New("redis: nil"),
			itemToAdd:   CartItem{ProductID: "P1", Quantity: 2, Price: 10.0},
			validateCart: func(t *testing.T, cart Cart) {
				assert.Equal(t, userID, cart.UserID)
				assert.Len(t, cart.Items, 1)
				assert.Equal(t, 2, cart.Items[0].Quantity)
			},
		},
		{
			name: "Éxito: El producto ya existe en el carrito, se suma la cantidad",
			initialCart: Cart{
				UserID: userID,
				Items: []CartItem{{ProductID: "P1", Quantity: 2, Price: 10.0}},
			},
			initialErr: nil,
			itemToAdd:  CartItem{ProductID: "P1", Quantity: 3, Price: 10.0},
			validateCart: func(t *testing.T, cart Cart) {
				assert.Len(t, cart.Items, 1)
				assert.Equal(t, 5, cart.Items[0].Quantity)
			},
		},
		{
			name:          "Error Borde: El canal persistCh está lleno",
			initialCart:   Cart{UserID: userID, Items: []CartItem{}},
			initialErr:    nil,
			itemToAdd:     CartItem{ProductID: "P1", Quantity: 1},
			simulateBlock: true,
		},
		{
			name:          "Error: cacheRepo.SaveCart falla",
			initialCart:   Cart{UserID: userID},
			initialErr:    nil,
			itemToAdd:     CartItem{ProductID: "P1", Quantity: 1},
			mockSaveErr:   errors.New("redis error connection"),
			expectedError: "redis error connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := new(mockCacheRepo)
			db := new(mockDBRepo)
			catalog := new(mockCatalog)

			cache.On("GetCart", mock.Anything, userID).Return(tt.initialCart, tt.initialErr)

			if tt.mockSaveErr != nil {
				cache.On("SaveCart", mock.Anything, mock.AnythingOfType("cart.Cart")).Return(tt.mockSaveErr)
			} else {
				cache.On("SaveCart", mock.Anything, mock.AnythingOfType("cart.Cart")).Return(nil).Run(func(args mock.Arguments) {
					c := args.Get(1).(Cart)
					if tt.validateCart != nil {
						tt.validateCart(t, c)
					}
				})
			}

			// Para el patrón asíncrono en DB
			var wg sync.WaitGroup
			var blockDB chan struct{}
			if tt.mockSaveErr == nil && !tt.simulateBlock {
				wg.Add(1)
				db.On("PersistCart", mock.Anything, mock.AnythingOfType("cart.Cart")).Return(nil).Run(func(args mock.Arguments) {
					wg.Done()
				})
			} else if tt.simulateBlock {
				blockDB = make(chan struct{})
				// Simulamos que la DB está colgada, para que el worker espere y el buffer se llene instantáneamente
				db.On("PersistCart", mock.Anything, mock.AnythingOfType("cart.Cart")).Run(func(args mock.Arguments) {
					<-blockDB
				}).Return(nil)
			}

			s := NewService(cache, db, catalog).(*cartService)

			// Simula canal saturado (Write-Behind demorado / bloqueado)
			if tt.simulateBlock {
				for i := 0; i < cap(s.persistCh); i++ {
					select {
					case s.persistCh <- Cart{}:
					default:
					}
				}
			}

			err := s.AddToCart(context.Background(), userID, tt.itemToAdd)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if tt.mockSaveErr == nil && !tt.simulateBlock {
				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()
				// Damos tiempo a la DB asíncrona a procesar
				select {
				case <-time.After(1 * time.Second):
					t.Logf("Timeout db.PersistCart in test")
				case <-done:
				}
			}

			// Desbloqueamos el worker para que drene la cola rápidamente y el Close() no se cuelgue
			if tt.simulateBlock {
				close(blockDB)
			}

			// Concurrencia segura: Drenar y cerrar para que la goroutina no quede "leaked" en el suite
			s.Close()
			cache.AssertExpectations(t)
		})
	}
}

// --- TESTS: GetValidatedCart ---

func TestService_GetValidatedCart(t *testing.T) {
	userID := "user-val-123"

	tests := []struct {
		name         string
		initialCart  Cart
		mockStock    map[string]int
		expectSave   bool
		validateCart func(t *testing.T, cart Cart)
	}{
		{
			name: "Éxito: El stock es suficiente, el carrito retorna intacto",
			initialCart: Cart{
				UserID: userID,
				Items:  []CartItem{{ProductID: "P1", Quantity: 2}},
			},
			mockStock:  map[string]int{"P1": 10},
			expectSave: false,
			validateCart: func(t *testing.T, cart Cart) {
				assert.Equal(t, 2, cart.Items[0].Quantity)
			},
		},
		{
			name: "Caso Borde (Ajuste): Usuario pidió 5, quedan 2. Ajuste a 2 y llama a SaveCart.",
			initialCart: Cart{
				UserID: userID,
				Items:  []CartItem{{ProductID: "P1", Quantity: 5}},
			},
			mockStock:  map[string]int{"P1": 2},
			expectSave: true,
			validateCart: func(t *testing.T, cart Cart) {
				assert.Equal(t, 2, cart.Items[0].Quantity)
			},
		},
		{
			name: "Caso Borde (Agotado): Stock es 0. Elimina item del carrito.",
			initialCart: Cart{
				UserID: userID,
				Items: []CartItem{
					{ProductID: "P1", Quantity: 2},
					{ProductID: "P2", Quantity: 1},
				},
			},
			mockStock:  map[string]int{"P1": 0, "P2": 10},
			expectSave: true,
			validateCart: func(t *testing.T, cart Cart) {
				assert.Len(t, cart.Items, 1)
				assert.Equal(t, "P2", cart.Items[0].ProductID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := new(mockCacheRepo)
			db := new(mockDBRepo)
			catalog := new(mockCatalog)

			cache.On("GetCart", mock.Anything, userID).Return(tt.initialCart, nil)

			var pids []string
			for _, item := range tt.initialCart.Items {
				pids = append(pids, item.ProductID)
			}
			catalog.On("CheckStock", mock.Anything, pids).Return(tt.mockStock, nil)

			var wg sync.WaitGroup
			if tt.expectSave {
				cache.On("SaveCart", mock.Anything, mock.AnythingOfType("cart.Cart")).Return(nil)
				
				wg.Add(1)
				db.On("PersistCart", mock.Anything, mock.AnythingOfType("cart.Cart")).Return(nil).Run(func(args mock.Arguments) {
					wg.Done()
				})
			}

			s := NewService(cache, db, catalog).(*cartService)

			validatedCart, err := s.GetValidatedCart(context.Background(), userID)
			assert.NoError(t, err)

			if tt.validateCart != nil {
				tt.validateCart(t, *validatedCart)
			}

			if tt.expectSave {
				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()
				<-done
			}

			s.Close()
			cache.AssertExpectations(t)
			catalog.AssertExpectations(t)
		})
	}
}
