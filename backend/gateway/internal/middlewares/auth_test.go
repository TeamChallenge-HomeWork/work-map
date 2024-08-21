package middlewares

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"workmap/gateway/internal/redis"
)

type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) CheckAccessToken(accessToken string) error {
	args := m.Called(accessToken)

	return args.Error(0)
}

type mockHandler struct {
	called bool
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
}

func TestCheckAuth(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		header         string
		expectedCode   int
		mockRedisError error
		handlerCalled  bool
	}{
		{
			name:           "Valid Token",
			header:         "Bearer valid-token-string",
			expectedCode:   http.StatusOK,
			mockRedisError: nil,
			handlerCalled:  true,
		},
		{
			name:           "Token does not exist",
			header:         "Bearer token-not-exist",
			expectedCode:   http.StatusUnauthorized,
			mockRedisError: errors.New("unauthorized"),
			handlerCalled:  false,
		},
		{
			name:           "Empty Authorization Header",
			header:         "",
			expectedCode:   http.StatusUnauthorized,
			mockRedisError: nil,
			handlerCalled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedis := new(MockRedis)
			var mockRedisStore store.TokenChecker = mockRedis

			middleware := &Middleware{
				logger: logger,
				redis:  mockRedisStore,
			}

			mockRedis.On("CheckAccessToken", mock.Anything).Return(tt.mockRedisError)

			req, err := http.NewRequest("", "", bytes.NewBuffer([]byte{}))
			if err != nil {
				t.Error(err)
			}
			req.Header.Set("Authorization", tt.header)

			w := httptest.NewRecorder()

			handler := &mockHandler{}
			middleware.CheckAuth(handler.ServeHTTP)(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Equal(t, tt.handlerCalled, handler.called)
		})
	}
}
