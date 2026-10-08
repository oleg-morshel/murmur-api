package testutil

import (
	"context"
	"io"
	"log/slog"

	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

func NewContext() context.Context {
	log := &logger.Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	return logger.WithContext(context.Background(), log)
}
