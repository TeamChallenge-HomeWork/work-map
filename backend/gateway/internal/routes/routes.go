package routes

import (
	"go.uber.org/zap"
	"net/http"
	"workmap/gateway/internal/handlers"
	"workmap/gateway/internal/middlewares"
)

type Config struct {
	Logger     *zap.Logger
	Handler    *handlers.Handler
	Middleware *middlewares.Middleware
}

type Router struct {
	handler    *handlers.Handler
	middleware *middlewares.Middleware
}

func New(cfg *Config) *Router {
	return &Router{
		handler:    cfg.Handler,
		middleware: cfg.Middleware,
	}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	h, m := r.handler, r.middleware

	mux.HandleFunc("OPTIONS /", m.EnableCORS(preflight))

	mux.HandleFunc("POST /user/register", m.EnableCORS(h.UserRegister))
	mux.HandleFunc("POST /user/login", m.EnableCORS(h.UserLogin)) // TODO add CheckAuth middleware and (?)redirect or delegate to FrontEnd
	mux.HandleFunc("POST /user/refreshtoken", m.EnableCORS(h.UserRefreshToken))
}

func preflight(w http.ResponseWriter, r *http.Request) {}
