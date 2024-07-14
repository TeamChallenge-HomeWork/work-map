package middlewares

import (
	"go.uber.org/zap"
	"workmap/gateway/internal/cache"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

type Config struct {
	Logger *zap.Logger
	Auth   pb.AuthServiceClient
	Redis  cache.Redis
}

type Middleware struct {
	logger *zap.Logger
	auth   pb.AuthServiceClient
	redis  cache.Redis
}

func New(cfg *Config) *Middleware {
	return &Middleware{
		logger: cfg.Logger,
		auth:   cfg.Auth,
		redis:  cfg.Redis,
	}
}
