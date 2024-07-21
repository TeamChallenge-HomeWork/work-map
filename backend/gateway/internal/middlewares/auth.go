package middlewares

import (
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func (m *Middleware) CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		m.logger.Debug(tokenString)
		token := strings.TrimPrefix(tokenString, "Bearer ")

		res, err := m.redis.Client.Get("access_token:" + token).Result()
		if err != nil {
			m.logger.Error("token not found", zap.Error(err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		m.logger.Debug("result", zap.String("res", res))

		next.ServeHTTP(w, r)
	}
}
