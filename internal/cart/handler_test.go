package cart

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"profit-ecommerce/internal/api/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCK SERVICE ---
type mockService struct {
	mock.Mock
}

func (m *mockService) AddToCart(ctx context.Context, userID string, item CartItem) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}

func (m *mockService) GetValidatedCart(ctx context.Context, userID string) (*Cart, error) {
	args := m.Called(ctx, userID)
	if cart := args.Get(0); cart != nil {
		return cart.(*Cart), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// --- TESTS: HandleAddToCart ---
func TestCartHandler_HandleAddToCart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		userIDHeader   string
		requestBody    interface{} // string, o struct para serializar automáticamente
		setupMock      func(m *mockService)
		expectedStatus int
	}{
		{
			name:           "Error 401: Falta el header X-User-ID",
			userIDHeader:   "",
			requestBody:    CartItem{ProductID: "P1", Quantity: 1, Price: 10.0},
			setupMock:      func(m *mockService) {}, // No debe ser llamado
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Error 400: Body con JSON inválido",
			userIDHeader:   "user-123",
			requestBody:    "json-invalido-o-incompleto",
			setupMock:      func(m *mockService) {}, // No debe ser llamado
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "Error 500: El servicio mockeado retorna un error",
			userIDHeader: "user-123",
			requestBody:  CartItem{ProductID: "P1", Quantity: 1, Price: 10.0},
			setupMock: func(m *mockService) {
				m.On("AddToCart", mock.Anything, "user-123", mock.AnythingOfType("cart.CartItem")).Return(errors.New("db timeout"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:         "Éxito 201: Header y body correctos, el servicio no retorna error",
			userIDHeader: "user-123",
			requestBody:  CartItem{ProductID: "P1", Quantity: 2, Price: 15.0},
			setupMock: func(m *mockService) {
				m.On("AddToCart", mock.Anything, "user-123", CartItem{ProductID: "P1", Quantity: 2, Price: 15.0}).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := new(mockService)
			tt.setupMock(mSvc)

			handler := NewCartHandler(mSvc)

			var reqBody []byte
			switch v := tt.requestBody.(type) {
			case string:
				reqBody = []byte(v)
			default:
				reqBody, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(reqBody))
			if tt.userIDHeader != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userIDHeader)
				req = req.WithContext(ctx)
			}

			// Recorder para capturar la respuesta
			rr := httptest.NewRecorder()
			handler.HandleAddToCart(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mSvc.AssertExpectations(t)
		})
	}
}

// --- TESTS: HandleGetCart ---
func TestCartHandler_HandleGetCart(t *testing.T) {
	t.Parallel()

	userID := "user-123"
	mockValidCart := &Cart{
		UserID: userID,
		Items: []CartItem{
			{ProductID: "P1", Quantity: 2, Price: 20.0},
		},
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		userIDHeader   string
		setupMock      func(m *mockService)
		expectedStatus int
		validateRes    func(t *testing.T, res *httptest.ResponseRecorder)
	}{
		{
			name:           "Error 401: Falta el header X-User-ID",
			userIDHeader:   "",
			setupMock:      func(m *mockService) {}, // No llamado
			expectedStatus: http.StatusUnauthorized,
			validateRes:    nil,
		},
		{
			name:         "Error 500: El servicio retorna error",
			userIDHeader: userID,
			setupMock: func(m *mockService) {
				m.On("GetValidatedCart", mock.Anything, userID).Return(nil, errors.New("timeout redis"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateRes:    nil,
		},
		{
			name:         "Éxito 200: Retorna un carrito válido decodificando el JSON",
			userIDHeader: userID,
			setupMock: func(m *mockService) {
				m.On("GetValidatedCart", mock.Anything, userID).Return(mockValidCart, nil)
			},
			expectedStatus: http.StatusOK,
			// Verifica que el json output sea correcto
			validateRes: func(t *testing.T, res *httptest.ResponseRecorder) {
				var responseCart Cart
				err := json.NewDecoder(res.Body).Decode(&responseCart)
				assert.NoError(t, err)

				assert.Equal(t, userID, responseCart.UserID)
				assert.Len(t, responseCart.Items, 1)
				assert.Equal(t, "P1", responseCart.Items[0].ProductID)
				assert.Equal(t, 2, responseCart.Items[0].Quantity)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := new(mockService)
			tt.setupMock(mSvc)

			handler := NewCartHandler(mSvc)

			req := httptest.NewRequest(http.MethodGet, "/cart", nil)
			if tt.userIDHeader != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userIDHeader)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.HandleGetCart(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.validateRes != nil {
				tt.validateRes(t, rr)
			}

			mSvc.AssertExpectations(t)
		})
	}
}
