package middlewares

import (
	"go.uber.org/zap"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/store"
)

type Config struct {
	Logger *zap.Logger
	Auth   pb.AuthServiceClient
	Redis  store.Redis
}

type Middleware struct {
	logger *zap.Logger
	auth   pb.AuthServiceClient
	redis  store.Redis
}

func New(cfg *Config) *Middleware {
	return &Middleware{
		logger: cfg.Logger,
		auth:   cfg.Auth,
		redis:  cfg.Redis,
	}
}
