package middlewares

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
)

type data struct {
	AccessToken string `json:"accessToken"`
}

func (m *Middleware) CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var d data
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			m.logger.Error("failed to read body with access token", zap.Error(err))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		res, err := m.redis.Client.Get("access_token:" + d.AccessToken).Result()
		if err != nil {
			m.logger.Error("failed to WHAT?", zap.Error(err))
			http.Error(w, "unauthorized1", http.StatusUnauthorized)
			return
		}

		m.logger.Debug("result", zap.String("res", res))

		next.ServeHTTP(w, r)
	}
}
