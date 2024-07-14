package cache

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
)

type Config struct {
	Host     string
	Port     string
	Password string
}

type Redis struct {
	Client redis.Client
}

func New(cfg *Config) (Redis, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	var client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       0,
	})

	if client == nil {
		return Redis{}, errors.New("cannot run redis")
	}

	return Redis{
		Client: *client,
	}, nil
}
