package store

import (
	"errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectedError error
	}{
		{
			name:          "token exist",
			token:         "token_exist",
			expectedError: nil,
		},
		{
			name:          "token does not exist",
			token:         "token_does_not_exist",
			expectedError: errors.New("unauthorized"),
		},
	}

	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	store := &RedisStore{client: rdb}

	err = s.Set("access_token:token_exist", "email")
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = store.GetAccessToken(tt.token)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestSaveAccessToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectedError error
	}{
		{
			name:          "valid token",
			token:         "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q",
			expectedError: nil,
		},
		{
			name:          "invalid token",
			token:         "invalid_token",
			expectedError: errors.New("cannot split the token string"),
		},
		{
			name:          "email does not exist",
			token:         "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.juifekcJQHuBlKy4TVqtKC6zv72Xnn03tXSZrD0-8YyGVYZ5Ayu3awuS01QTFwsbsEqTbAelB2_D81U6iOBLTA",
			expectedError: errors.New("email not found in the token"),
		},
	}

	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	store := &RedisStore{client: rdb}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = store.SaveAccessToken(tt.token)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestDeleteAccessToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectedError error
	}{
		{
			name:          "token exist",
			token:         "token_exist",
			expectedError: nil,
		},
		{
			name:          "token does not exist",
			token:         "token_does_not_exist",
			expectedError: nil,
		},
	}

	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	store := &RedisStore{client: rdb}

	err = s.Set("access_token:token_exist", "email")
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = store.DeleteAccessToken(tt.token)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
