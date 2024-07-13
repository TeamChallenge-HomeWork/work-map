package routes

import (
	"go.uber.org/zap"
	"net/http"
	"workmap/gateway/internal/handlers"
	"workmap/gateway/internal/middlewares"
)

type Config struct {
	Mux    *http.ServeMux
	Logger *zap.Logger
}

type Router struct {
	mux        *http.ServeMux
	handler    *handlers.Handler
	middleware *middlewares.Middleware
}

func New(cfg *Config) *Router {
	handler := handlers.New(&handlers.Config{})
	middleware := middlewares.New(&middlewares.Config{})

	return &Router{
		mux:        cfg.Mux,
		handler:    handler,
		middleware: middleware,
	}
}

func (r *Router) RegisterRouters() {

}
