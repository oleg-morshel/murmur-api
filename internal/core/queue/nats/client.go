package core_nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type Client struct {
	conn *nats.Conn
}

func NewClient(ctx context.Context, cfg Config, log *logger.Logger) (*Client, error) {
	conn, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("core_nats.NewClient: connect: %w", err)
	}

	log.Info("nats connected", slog.String("url", cfg.URL))

	return &Client{conn: conn}, nil
}

func (c *Client) Publish(subject string, data []byte) error {
	if err := c.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("core_nats.Client.Publish: %w", err)
	}
	return nil
}

func (c *Client) Subscribe(subject string, handler func(data []byte)) (*nats.Subscription, error) {
	sub, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	if err != nil {
		return nil, fmt.Errorf("core_nats.Client.Subscribe: %w", err)
	}
	return sub, nil
}

func (c *Client) Close() {
	c.conn.Close()
}
