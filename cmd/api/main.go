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
	core_http_middleware "github.com/oleg-morshel/murmur-api/internal/core/transport/http/middleware"
	core_http_server "github.com/oleg-morshel/murmur-api/internal/core/transport/http/server"
	auth_postgres "github.com/oleg-morshel/murmur-api/internal/features/auth/repository/postgres"
	auth_service "github.com/oleg-morshel/murmur-api/internal/features/auth/service"
	auth_transport_http "github.com/oleg-morshel/murmur-api/internal/features/auth/transport/http"
	posts_postgres "github.com/oleg-morshel/murmur-api/internal/features/posts/repository/postgres"
	posts_service "github.com/oleg-morshel/murmur-api/internal/features/posts/service"
	posts_transport_http "github.com/oleg-morshel/murmur-api/internal/features/posts/transport/http"
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

	log.Info("pool timeout", slog.Duration("timeout", pool.OpTimeout()))

	log.Debug("initializing feature", slog.String("feature", "auth"))
	authRepository := auth_postgres.NewAuthRepository(pool)
	tokenRepository := auth_postgres.NewTokenRepository(pool)
	authService := auth_service.NewService(
		authRepository,
		tokenRepository,
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTTL,
		cfg.Auth.RefreshTTL,
	)
	authTransportHttp := auth_transport_http.NewAuthHTTPHandler(authService)

	log.Debug("initializing feature", slog.String("feature", "posts"))
	postRepository := posts_postgres.NewPostRepository(pool)
	postService := posts_service.NewPostService(postRepository)
	postsTransportHttp := posts_transport_http.NewPostsHTTPHandler(postService)

	log.Debug("initializing HTTP server")

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		log,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(log),
		core_http_middleware.Panic(),
		core_http_middleware.Trace(),
	)

	apiVersionRouter := core_http_server.NewAPIVersionRouter(
		core_http_server.ApiVersion1,
		auth_transport_http.AuthMiddleware(authService),
	)
	apiVersionRouter.RegisterRoutes(authTransportHttp.Routes()...)
	apiVersionRouter.RegisterRoutes(postsTransportHttp.Routes()...)

	httpServer.RegisterApiRouters(apiVersionRouter)

	log.Info(">>> murmur api STARTED")
	if err := httpServer.Run(ctx); err != nil {
		log.Error("HTTP server run error", slog.Any("error", err))
	}
	log.Info(">>> murmur api STOPPED")
}
