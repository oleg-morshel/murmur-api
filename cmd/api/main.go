package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/config"
	core_redis "github.com/oleg-morshel/murmur-api/internal/core/cache/redis"
	core_nats "github.com/oleg-morshel/murmur-api/internal/core/queue/nats"
	core_postgres_pool "github.com/oleg-morshel/murmur-api/internal/core/repository/postgres/pool"
	core_http_middleware "github.com/oleg-morshel/murmur-api/internal/core/transport/http/middleware"
	core_http_server "github.com/oleg-morshel/murmur-api/internal/core/transport/http/server"
	auth_postgres "github.com/oleg-morshel/murmur-api/internal/features/auth/repository/postgres"
	auth_service "github.com/oleg-morshel/murmur-api/internal/features/auth/service"
	auth_transport_http "github.com/oleg-morshel/murmur-api/internal/features/auth/transport/http"
	polls_postgres "github.com/oleg-morshel/murmur-api/internal/features/polls/repository/postgres"
	polls_service "github.com/oleg-morshel/murmur-api/internal/features/polls/service"
	polls_transport_http "github.com/oleg-morshel/murmur-api/internal/features/polls/transport/http"
	posts_cache "github.com/oleg-morshel/murmur-api/internal/features/posts/cache"
	posts_postgres "github.com/oleg-morshel/murmur-api/internal/features/posts/repository/postgres"
	posts_service "github.com/oleg-morshel/murmur-api/internal/features/posts/service"
	posts_transport_http "github.com/oleg-morshel/murmur-api/internal/features/posts/transport/http"
	"github.com/oleg-morshel/murmur-api/internal/grpcserver"
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

	redisClient, err := core_redis.NewClient(ctx, core_redis.NewConfigMust(), log)
	if err != nil {
		log.Error("failed to init redis", slog.Any("error", err))
		os.Exit(1)
	}
	defer redisClient.Close()

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

	natsClient, err := core_nats.NewClient(ctx, core_nats.NewConfigMust(), log)
	if err != nil {
		log.Error("failed to init nats", slog.Any("error", err))
		os.Exit(1)
	}
	defer natsClient.Close()

	log.Debug("initializing feature", slog.String("feature", "polls"))

	pollRepository := polls_postgres.NewPollRepository(pool)
	pollService := polls_service.NewPollService(pollRepository, natsClient)
	pollsTransportHttp := polls_transport_http.NewPollsHTTPHandler(pollService)

	log.Debug("initializing feature", slog.String("feature", "posts"))

	postRepository := posts_postgres.NewPostRepository(pool)
	postCache := posts_cache.NewRedisCache(redisClient.RDB())
	rateLimiter := posts_cache.NewRateLimiter(redisClient.RDB(), 10, time.Minute)
	postService := posts_service.NewPostService(postRepository, postCache, rateLimiter, pollService, natsClient)
	postsTransportHttp := posts_transport_http.NewPostsHTTPHandler(postService)

	log.Debug("initializing gRPC server")

	grpcSrv := grpcserver.New(authRepository, authService, log)
	go func() {
		if err := grpcSrv.Serve(cfg.GRPC.Addr); err != nil {
			log.Error("gRPC server run error", slog.Any("error", err))
			stop()
		}
	}()

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
	apiVersionRouter.RegisterRoutes(pollsTransportHttp.Routes()...)

	httpServer.RegisterApiRouters(apiVersionRouter)

	log.Info(">>> murmur api STARTED")
	if err := httpServer.Run(ctx); err != nil {
		log.Error("HTTP server run error", slog.Any("error", err))
	}
	grpcSrv.GracefulStop()

	log.Info(">>> murmur api STOPPED")
}
