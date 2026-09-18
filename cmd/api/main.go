package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/oleg-morshel/murmur-api/internal/config"
	core_postgres_pool "github.com/oleg-morshel/murmur-api/internal/core/repository/postgres/pool"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log, err := logger.New(logger.Config{
		Level:  cfg.Logger.Level,
		Folder: cfg.Logger.Folder,
	})
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Debug("config and logger initialized", slog.String("env", cfg.Env))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := core_postgres_pool.NewConnectionPool(ctx, core_postgres_pool.NewConfigMust(), log)

	if err != nil {
		log.Error("failed to init connection pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("murmur api started")

	<-ctx.Done()
	log.Info("shutting down gracefully...")
}
