package middlewares

import (
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func (m *Middleware) CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			m.logger.Error("token not found")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		m.logger.Debug(tokenString)
		token := strings.TrimPrefix(tokenString, "Bearer ")

		err := m.redis.CheckAccessToken(token)
		if err != nil {
			m.logger.Error("token not found", zap.Error(err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
