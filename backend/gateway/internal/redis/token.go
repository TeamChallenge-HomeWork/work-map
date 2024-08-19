package store

import (
	"errors"
	"github.com/go-redis/redis"
	"workmap/gateway/internal/pkg/token"
)

type TokenStore interface {
	TokenGetter
	TokenSetter
	TokenDeleter
}

type TokenGetter interface {
	GetAccessToken(accessToken string) error
}

type TokenSetter interface {
	SaveAccessToken(accessToken string) error
}

type TokenDeleter interface {
	DeleteAccessToken(accessToken string) error
}

func (r *RedisStore) GetAccessToken(accessToken string) error {
	res := r.client.Get("access_token:" + accessToken)
	if res.Err() != nil {
		if res.Err() == redis.Nil {
			return errors.New("unauthorized")
		}

		return res.Err()
	}

	return nil
}

func (r *RedisStore) SaveAccessToken(accessToken string) error {
	extractor := token.AccessTokenExtractor{}
	ttl, err := extractor.ExtractTTL(accessToken)
	if err != nil {
		return err
	}

	email, err := extractor.ExtractEmail(accessToken)
	if err != nil {
		return err
	}

	res := r.client.Set("access_token:"+accessToken, email, ttl)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (r *RedisStore) DeleteAccessToken(accessToken string) error {
	res := r.client.Del("access_token:" + accessToken)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}
