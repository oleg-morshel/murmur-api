package core_redis

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     string `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

func NewConfigMust() Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic(fmt.Sprintf("redis config: %v", err))
	}
	return cfg
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
