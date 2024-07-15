package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"workmap/gateway/config"
	"workmap/gateway/logger"
)

func main() {
	log := logger.New()
	cfg := config.New(log)
	services := cfg.NewServices(log)
	server := services.Server

	server.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.ShutDown(ctx)
}
