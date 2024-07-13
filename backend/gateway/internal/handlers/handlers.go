package handlers

import (
	"go.uber.org/zap"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

type Config struct {
	Logger *zap.Logger
	Auth   pb.AuthServiceClient
}

type Handler struct {
	logger *zap.Logger
	auth   pb.AuthServiceClient
}

func New(cfg *Config) *Handler {
	return &Handler{
		logger: cfg.Logger,
		auth:   cfg.Auth,
	}
}
