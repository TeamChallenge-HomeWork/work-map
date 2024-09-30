package logger

import (
	"go.uber.org/zap"
	"log"
)

func New() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("failed to init zap logger")
	}

	return logger
}
