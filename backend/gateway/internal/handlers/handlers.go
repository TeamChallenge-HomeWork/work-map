package handlers

import (
	"go.uber.org/zap"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/redis"
)

type Config struct {
	Logger     *zap.Logger
	Auth       pb.AuthServiceClient
	TokenStore store.TokenStore
}

type Handler struct {
	logger     *zap.Logger
	auth       pb.AuthServiceClient
	tokenStore store.TokenStore
}

func New(cfg *Config) *Handler {
	return &Handler{
		logger:     cfg.Logger,
		auth:       cfg.Auth,
		tokenStore: cfg.TokenStore,
	}
}
