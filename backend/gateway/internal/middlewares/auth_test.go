package middlewares_test

import (
	"bytes"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"workmap/gateway/internal/middlewares"
	"workmap/gateway/internal/store"
)

type MockRedisClient struct {
	mock.Mock
}

type RedisClient interface {
	Get(key string) (string, error)
}

type mockHandler struct {
	called bool
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
}

type data struct {
	AccessToken string `json:"accessToken"`
}

func TestCheckAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name          string
		header        string
		setupMock     func(m *MockRedisClient)
		expectedCode  int
		handlerCalled bool
	}{
		{
			name:          "Valid Token",
			header:        "Bearer valid-token",
			expectedCode:  http.StatusOK,
			handlerCalled: true,
		},
		{
			name:          "Invalid Token",
			header:        "Bearer not-valid-token",
			expectedCode:  http.StatusUnauthorized,
			handlerCalled: false,
		},
		{
			name:          "Empty Authorization Header",
			header:        "",
			expectedCode:  http.StatusUnauthorized,
			handlerCalled: false,
		},
	}

	server, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	r := &store.Redis{Client: *rc}

	err := r.Client.Set("access_token:valid-token", "test@token.com", time.Duration(10)*time.Second).Err()
	if err != nil {
		t.Fatal(err)
	}

	cfg := &middlewares.Config{
		Logger: logger,
		Redis:  *r,
	}
	middleware := middlewares.New(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte{}))
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			w := httptest.NewRecorder()

			handler := &mockHandler{}
			middleware.CheckAuth(handler.ServeHTTP)(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Equal(t, tt.handlerCalled, handler.called)
		})
	}
}
