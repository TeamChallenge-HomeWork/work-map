package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"workmap/gateway/internal/store"
)

type mockHandler struct {
	called bool
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
}

func TestCheckAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name          string
		body          interface{}
		setupRedis    func(mr *miniredis.Miniredis)
		expectedCode  int
		handlerCalled bool
	}{
		{
			name: "Valid Token",
			body: data{AccessToken: "valid-token"},
			setupRedis: func(mr *miniredis.Miniredis) {
				mr.Set("access_token:valid-token", "1")
			},
			expectedCode:  http.StatusOK,
			handlerCalled: true,
		},
		{
			name: "Invalid Token",
			body: data{AccessToken: "invalid-token"},
			setupRedis: func(mr *miniredis.Miniredis) {
				// No setup needed for invalid token
			},
			expectedCode:  http.StatusUnauthorized,
			handlerCalled: false,
		},
		{
			name: "Empty Body",
			body: nil,
			setupRedis: func(mr *miniredis.Miniredis) {
				// No setup needed for empty body
			},
			expectedCode:  http.StatusUnauthorized,
			handlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, err := miniredis.Run()
			if err != nil {
				t.Fatalf("failed to start miniredis: %v", err)
			}
			defer mr.Close()

			tt.setupRedis(mr)

			r, err := store.NewRedis(&store.RedisConfig{
				Host:     "100.104.232.63",
				Port:     "6366",
				Password: "password",
			})
			if err != nil {
				t.Fatal(err)
			}
			cfg := &Config{
				Logger: logger,
				Redis:  r,
			}
			middleware := New(cfg)

			var reqBody []byte
			if tt.body != nil {
				reqBody, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			handler := &mockHandler{}
			middleware.CheckAuth(handler.ServeHTTP)(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Equal(t, tt.handlerCalled, handler.called)
		})
	}
}
