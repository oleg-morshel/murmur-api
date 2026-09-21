package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env    string `envconfig:"ENV" default:"local"`
	Logger Logger
	Auth   AuthConfig
}

type Logger struct {
	Level  string `envconfig:"LOGGER_LEVEL" default:"debug"`
	Folder string `envconfig:"LOGGER_FOLDER" default:"./logs"`
}

type AuthConfig struct {
	JWTSecret  string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTTL  time.Duration `envconfig:"ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `envconfig:"REFRESH_TTL" default:"720h"`
}

func MustLoad() *Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic(fmt.Sprintf("config.MustLoad: %v", err))
	}
	return &cfg
}
