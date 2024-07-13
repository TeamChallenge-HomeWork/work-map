package server

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/routes"
)

type Config struct {
	Port   string
	Logger *zap.Logger
	Auth   pb.AuthServiceClient
}

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func New(cfg *Config) *Server {
	mux := http.NewServeMux()

	router := routes.New(&routes.Config{
		Mux:    mux,
		Logger: cfg.Logger,
		Auth:   cfg.Auth,
	})
	router.RegisterRouters()

	addr := fmt.Sprintf(":%s", cfg.Port)
	srvr := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &Server{
		httpServer: srvr,
		logger:     cfg.Logger,
	}
}

func (s *Server) Run() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.logger.Fatal("failed to listen server", zap.String("address", s.httpServer.Addr), zap.Error(err))
		}
	}()
	s.logger.Info("server is ready to handle requests", zap.String("address", s.httpServer.Addr))
}

func (s *Server) ShutDown(ctx context.Context) {
	s.logger.Debug("Shutting down gracefully, press Ctrl+C again to force")
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Fatal("Server forced to shutdown: %v", zap.Error(err))
	}
	s.logger.Debug("Server stopped")
}
