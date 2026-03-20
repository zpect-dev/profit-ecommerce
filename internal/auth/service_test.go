package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCK CACHE REPOSITORIO ---

type mockAuthCacheRepo struct {
	mock.Mock
}

func (m *mockAuthCacheRepo) SaveSession(ctx context.Context, session Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockAuthCacheRepo) GetSession(ctx context.Context, token string) (Session, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(Session), args.Error(1)
}

func (m *mockAuthCacheRepo) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// --- MOCK CLIENT REPOSITORIO (PostgreSQL) ---

type mockClientRepo struct {
	mock.Mock
}

func (m *mockClientRepo) FindClientByID(ctx context.Context, coCli string) (ClientRow, error) {
	args := m.Called(ctx, coCli)
	return args.Get(0).(ClientRow), args.Error(1)
}

// --- TESTS: Login ---

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		mockSetup     func(cache *mockAuthCacheRepo, clientDB *mockClientRepo)
		expectToken   bool
		expectedError string
	}{
		{
			name:     "Éxito: Credenciales válidas, cliente activo en BD",
			username: "CLI001",
			password: "CLI001",
			mockSetup: func(cache *mockAuthCacheRepo, clientDB *mockClientRepo) {
				clientDB.On("FindClientByID", mock.Anything, "CLI001").Return(ClientRow{
					CoCli:    "CLI001",
					Tipo:     "A",
					CliDes:   "Cliente Demo SA",
					Inactivo: false,
					Login:    5000.50,
				}, nil).Once()

				cache.On("SaveSession", mock.Anything, mock.MatchedBy(func(s Session) bool {
					return s.UserID == "CLI001" &&
						s.CliDes == "Cliente Demo SA" &&
						s.Tipo == "A" &&
						s.MontCre == 5000.50 &&
						s.Token != "" &&
						s.ExpiresAt.After(time.Now())
				})).Return(nil).Once()
			},
			expectToken:   true,
			expectedError: "",
		},
		{
			name:     "Error: Credenciales no coinciden (username != password)",
			username: "CLI001",
			password: "wrong-pass",
			mockSetup: func(cache *mockAuthCacheRepo, clientDB *mockClientRepo) {
				// No se llama a ningún repo si las credenciales fallan
			},
			expectToken:   false,
			expectedError: "unauthorized",
		},
		{
			name:     "Error: Cliente no encontrado en BD",
			username: "NOEXIST",
			password: "NOEXIST",
			mockSetup: func(cache *mockAuthCacheRepo, clientDB *mockClientRepo) {
				clientDB.On("FindClientByID", mock.Anything, "NOEXIST").Return(ClientRow{}, errors.New("client not found")).Once()
			},
			expectToken:   false,
			expectedError: "unauthorized",
		},
		{
			name:     "Error: Cliente existe pero está Inactivo",
			username: "CLI002",
			password: "CLI002",
			mockSetup: func(cache *mockAuthCacheRepo, clientDB *mockClientRepo) {
				clientDB.On("FindClientByID", mock.Anything, "CLI002").Return(ClientRow{
					CoCli:    "CLI002",
					Tipo:     "B",
					CliDes:   "Cliente Inactivo",
					Inactivo: true,
					Login:    0,
				}, nil).Once()
			},
			expectToken:   false,
			expectedError: "unauthorized",
		},
		{
			name:     "Error: Redis falla al guardar sesión",
			username: "CLI003",
			password: "CLI003",
			mockSetup: func(cache *mockAuthCacheRepo, clientDB *mockClientRepo) {
				clientDB.On("FindClientByID", mock.Anything, "CLI003").Return(ClientRow{
					CoCli:    "CLI003",
					Tipo:     "A",
					CliDes:   "Cliente Redis Fail",
					Inactivo: false,
					Login:    1000,
				}, nil).Once()

				cache.On("SaveSession", mock.Anything, mock.Anything).Return(errors.New("redis: connection refused")).Once()
			},
			expectToken:   false,
			expectedError: "redis: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := new(mockAuthCacheRepo)
			clientDB := new(mockClientRepo)
			tt.mockSetup(cache, clientDB)

			svc := NewAuthService(cache, clientDB)
			resp, err := svc.Login(context.Background(), tt.username, tt.password)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.expectToken {
					assert.NotEmpty(t, resp.Token)
					assert.Equal(t, "CLI001", resp.UserID)
					assert.Equal(t, "Cliente Demo SA", resp.CliDes)
					assert.Equal(t, "A", resp.Tipo)
					assert.Equal(t, 5000.50, resp.MontCre)
				}
			}
			cache.AssertExpectations(t)
			clientDB.AssertExpectations(t)
		})
	}
}

// --- TESTS: ValidateToken ---

func TestAuthService_ValidateToken(t *testing.T) {
	validToken := "valid-token-hash-123"

	tests := []struct {
		name           string
		token          string
		mockSetup      func(cache *mockAuthCacheRepo)
		expectedUserID string
		expectedError  string
	}{
		{
			name:  "Éxito: Token válido extrae el UserID correcto (No Expirado)",
			token: validToken,
			mockSetup: func(cache *mockAuthCacheRepo) {
				cache.On("GetSession", mock.Anything, validToken).Return(Session{
					Token:     validToken,
					UserID:    "user-123",
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil).Once()
			},
			expectedUserID: "user-123",
			expectedError:  "",
		},
		{
			name:  "Error: Token inexistente/inválido (Redis devuelve nil) -> Unauthorized",
			token: "invalid-token",
			mockSetup: func(cache *mockAuthCacheRepo) {
				cache.On("GetSession", mock.Anything, "invalid-token").Return(Session{}, errors.New("redis: nil")).Once()
			},
			expectedUserID: "",
			expectedError:  "unauthorized",
		},
		{
			name:  "Error: Token expirado -> Unauthorized",
			token: "expired-token",
			mockSetup: func(cache *mockAuthCacheRepo) {
				cache.On("GetSession", mock.Anything, "expired-token").Return(Session{
					Token:     "expired-token",
					UserID:    "user-123",
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}, nil).Once()
			},
			expectedUserID: "",
			expectedError:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := new(mockAuthCacheRepo)
			clientDB := new(mockClientRepo) // no se usa en ValidateToken, pero requerido por constructor
			tt.mockSetup(cache)

			svc := NewAuthService(cache, clientDB)
			userID, err := svc.ValidateToken(context.Background(), tt.token)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserID, userID)
			}
			cache.AssertExpectations(t)
		})
	}
}
