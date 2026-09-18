package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env    string `envconfig:"ENV" default:"local"`
	Logger Logger
}

type Logger struct {
	Level  string `envconfig:"LOGGER_LEVEL" default:"debug"`
	Folder string `envconfig:"LOGGER_FOLDER" default:"./logs"`
}

func MustLoad() *Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic(fmt.Sprintf("config.MustLoad: %v", err))
	}
	return &cfg
}
