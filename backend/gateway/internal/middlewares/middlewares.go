package middlewares

type Config struct{}

type Middleware struct{}

func New(cfg *Config) *Middleware {
	return &Middleware{}
}
