package middlewares

import (
	"go.uber.org/zap"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/redis"
)

type Config struct {
	Logger *zap.Logger
	Auth   pb.AuthServiceClient
	Redis  store.TokenChecker
}

type Middleware struct {
	logger *zap.Logger
	auth   pb.AuthServiceClient
	redis  store.TokenChecker
}

func New(cfg *Config) *Middleware {
	return &Middleware{
		logger: cfg.Logger,
		auth:   cfg.Auth,
		redis:  cfg.Redis,
	}
}
