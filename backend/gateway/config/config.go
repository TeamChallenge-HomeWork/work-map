package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"workmap/gateway/internal/gapi"
	"workmap/gateway/internal/handlers"
	"workmap/gateway/internal/middlewares"
	"workmap/gateway/internal/redis"
	"workmap/gateway/internal/routes"
	"workmap/gateway/internal/server"
)

type (
	Config struct {
		Port        string      `mapstructure:"PORT"`
		AuthService AuthService `mapstructure:",squash"`
		Redis       Redis       `mapstructure:",squash"`
	}

	AuthService struct {
		Host string `mapstructure:"AUTH_SERVICE_HOST"`
		Port string `mapstructure:"AUTH_SERVICE_PORT"`
	}

	Redis struct {
		Host     string `mapstructure:"REDIS_HOST"`
		Port     string `mapstructure:"REDIS_PORT"`
		Password string `mapstructure:"REDIS_PASSWORD"`
	}
)

func New(logger *zap.Logger) *Config {
	var cfg Config

	v := viper.New()
	v.SetConfigType("env")
	v.AddConfigPath(".")    // path for local development
	v.AddConfigPath("/app") // path for container
	v.SetConfigName(".env")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
	}

	if err := v.Unmarshal(&cfg); err != nil {
		logger.Fatal("failed to unmarshal into config struct", zap.Error(err))
	}

	return &cfg
}

type Services struct {
	Server *server.Server
}

func (cfg *Config) NewServices(logger *zap.Logger) *Services {
	auth, err := gapi.NewAuthService(&gapi.AuthConfig{
		Host: cfg.AuthService.Host,
		Port: cfg.AuthService.Port,
	})
	if err != nil { // TODO delete this
		logger.Fatal("auth service err", zap.Error(err))
	}

	redis, err := store.NewRedis(&store.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
	})
	if err != nil {
		logger.Fatal("failed connection to redis", zap.Error(err))
	}

	h := handlers.New(&handlers.Config{
		Logger:     logger,
		Auth:       auth,
		TokenStore: &redis,
	})

	m := middlewares.New(&middlewares.Config{
		Logger: logger,
		Auth:   auth,
		Redis:  &redis,
	})

	r := routes.New(&routes.Config{
		Logger:     logger,
		Handler:    h,
		Middleware: m,
	})

	s := server.New(&server.Config{
		Port:   cfg.Port,
		Logger: logger,
		Router: r,
	})

	return &Services{
		Server: s,
	}
}
