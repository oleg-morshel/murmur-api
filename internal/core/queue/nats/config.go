package core_nats

import "os"

type Config struct {
	URL string
}

func NewConfigMust() Config {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "nats://localhost:4222"
	}

	return Config{URL: url}
}
