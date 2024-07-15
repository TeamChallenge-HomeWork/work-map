package store

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type Redis struct {
	Client redis.Client
}

func NewRedis(cfg *RedisConfig) (Redis, error) {
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
