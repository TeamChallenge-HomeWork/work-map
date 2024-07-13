package main

import (
	"workmap/gateway/config"
	"workmap/gateway/logger"
)

func main() {
	log := logger.New()
	cfg := config.New(log)
	_ = cfg
}
