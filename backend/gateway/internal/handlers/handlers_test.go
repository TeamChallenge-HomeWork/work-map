package handlers

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestNew(t *testing.T) {
	logger := zap.NewNop()
	mockAuthService := new(MockAuthServiceClient)
	mockRedis := new(MockRedis)

	cfg := &Config{
		Logger:     logger,
		Auth:       mockAuthService,
		TokenStore: mockRedis,
	}

	handler := New(cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, mockAuthService, handler.auth)
}
