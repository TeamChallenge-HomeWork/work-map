package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"workmap/gateway/internal/cache"
	"workmap/gateway/internal/gapi"
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

	// TODO refactor this shit because output is: config struct   {"": {"Port":"100.104.232.63","Port":"8080"}}
	logger.Debug("config struct", zap.Any("", cfg))
	return &cfg
}

type Services struct {
	Server *server.Server
}

func (cfg *Config) InitServices(logger *zap.Logger) *Services {
	auth, err := gapi.NewAuthService(&gapi.AuthConfig{
		Host: cfg.AuthService.Host,
		Port: cfg.AuthService.Port,
	})
	if err != nil { // TODO for what?
		logger.Fatal("failed to connect to auth service", zap.Error(err))
	}

	r, err := cache.New(&cache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
	})
	if err != nil {
		logger.Fatal("failed connection to redis client", zap.Error(err))
	}

	srvr := server.New(&server.Config{
		Port:   cfg.Port,
		Logger: logger,
		Auth:   auth,
		Redis:  r,
	})

	return &Services{
		Server: srvr,
	}
}
