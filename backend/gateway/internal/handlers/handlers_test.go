package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

type MockAuthServiceClient struct {
	mock.Mock
}

func TestNew(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockAuthService := new(MockAuthServiceClient)

	cfg := &Config{
		Logger: logger,
		Auth:   mockAuthService,
	}

	handler := New(cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, mockAuthService, handler.auth)
}
