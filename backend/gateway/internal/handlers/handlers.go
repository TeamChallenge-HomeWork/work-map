package handlers

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

type Handler struct {
	logger *zap.Logger
	auth   pb.AuthServiceClient
	redis  cache.Redis
}

func New(cfg *Config) *Handler {
	return &Handler{
		logger: cfg.Logger,
		auth:   cfg.Auth,
		redis:  cfg.Redis,
	}
}
