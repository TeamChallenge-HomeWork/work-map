package handlers

type Config struct{}

type Handler struct{}

func New(cfg *Config) *Handler {
	return &Handler{}
}
