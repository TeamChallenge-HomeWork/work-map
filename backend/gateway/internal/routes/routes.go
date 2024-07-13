package routes

import (
	"go.uber.org/zap"
	"net/http"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/handlers"
	"workmap/gateway/internal/middlewares"
)

type Config struct {
	Mux    *http.ServeMux
	Logger *zap.Logger
	Auth   pb.AuthServiceClient
}

type Router struct {
	mux        *http.ServeMux
	handler    *handlers.Handler
	middleware *middlewares.Middleware
}

func New(cfg *Config) *Router {
	handler := handlers.New(&handlers.Config{
		Logger: cfg.Logger,
		Auth:   cfg.Auth,
	})
	middleware := middlewares.New(&middlewares.Config{
		Logger: cfg.Logger,
		Auth:   cfg.Auth,
	})

	return &Router{
		mux:        cfg.Mux,
		handler:    handler,
		middleware: middleware,
	}
}

func (r *Router) RegisterRouters() {
	h, m := r.handler, r.middleware
	_ = h

	r.mux.HandleFunc("OPTIONS /", m.EnableCORS(preflight))
}

func preflight(w http.ResponseWriter, r *http.Request) {}
