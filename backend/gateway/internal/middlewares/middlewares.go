package middlewares

import (
	"go.uber.org/zap"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

type Config struct {
	Logger *zap.Logger
	Auth   pb.AuthServiceClient
}

type Middleware struct {
	logger *zap.Logger
	auth   pb.AuthServiceClient
}

func New(cfg *Config) *Middleware {
	return &Middleware{
		logger: cfg.Logger,
	}
}
