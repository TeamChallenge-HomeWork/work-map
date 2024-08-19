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

type RedisStore struct {
	client *redis.Client
}

func NewRedis(cfg *RedisConfig) (RedisStore, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	var client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       0,
	})
	if client == nil {
		return RedisStore{}, errors.New("cannot run redis")
	}

	err := client.Ping().Err()
	if err != nil {
		return RedisStore{}, errors.New("cannot ping to redis")
	}

	return RedisStore{
		client: client,
	}, nil
}
