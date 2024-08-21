package store

import (
	"errors"
	"github.com/go-redis/redis"
	"workmap/gateway/internal/pkg/token"
)

type TokenStore interface {
	TokenChecker
	SaveAccessToken(accessToken string) error
	DeleteAccessToken(accessToken string) error
}

type TokenChecker interface {
	CheckAccessToken(accessToken string) error
}

func (r *RedisStore) CheckAccessToken(accessToken string) error {
	err := r.client.Get("access_token:" + accessToken).Err()
	if err != nil {
		if err == redis.Nil {
			return errors.New("unauthorized")
		}
		return err
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

	err = r.client.Set("access_token:"+accessToken, email, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisStore) DeleteAccessToken(accessToken string) error {
	err := r.client.Del("access_token:" + accessToken).Err()
	if err != nil {
		return err
	}

	return nil
}
