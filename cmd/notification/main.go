package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oleg-morshel/murmur-api/internal/config"
	core_events "github.com/oleg-morshel/murmur-api/internal/core/events"
	core_nats "github.com/oleg-morshel/murmur-api/internal/core/queue/nats"
	"github.com/oleg-morshel/murmur-api/internal/grpcclient"
	"github.com/oleg-morshel/murmur-api/internal/ws"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
)

type postCreatedMessage struct {
	Type           string    `json:"type"`
	PostID         int64     `json:"post_id"`
	AuthorID       int64     `json:"author_id"`
	AuthorUsername string    `json:"author_username,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

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

	natsClient, err := core_nats.NewClient(ctx, core_nats.NewConfigMust(), log)
	if err != nil {
		log.Error("failed to init nats", slog.Any("error", err))
		os.Exit(1)
	}
	defer natsClient.Close()

	hub := ws.NewHub(log)
	go hub.Run()

	userClient, err := grpcclient.NewUserClient(cfg.GRPC.Target, []byte(cfg.Auth.JWTSecret))
	if err != nil {
		log.Error("failed to init grpc client", slog.Any("error", err))
		os.Exit(1)
	}
	defer userClient.Close()

	handlePostCreated := func(data []byte) {
		var event core_events.PostCreatedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Warn("notification: failed to unmarshal post.created", slog.Any("error", err))
			return
		}
		log.Info("notification: post created",
			slog.Int64("post_id", event.PostID),
			slog.Int64("author_id", event.AuthorID),
			slog.Time("timestamp", event.Timestamp),
		)

		username, err := userClient.GetUsername(ctx, event.AuthorID)
		if err != nil {
			log.Warn("notification: failed to get author via grpc",
				slog.Int64("author_id", event.AuthorID),
				slog.Any("error", err),
			)
		}

		msg, err := json.Marshal(postCreatedMessage{
			Type:           "post.created",
			PostID:         event.PostID,
			AuthorID:       event.AuthorID,
			AuthorUsername: username,
			Timestamp:      event.Timestamp,
		})
		if err != nil {
			log.Error("notification: failed to marshal ws message", slog.Any("error", err))
			return
		}

		hub.Broadcast(msg)
	}

	handlePostDeleted := func(data []byte) {
		var event core_events.PostDeletedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Warn("notification: failed to unmarshal post.deleted", slog.Any("error", err))
			return
		}
		log.Info("notification: post deleted",
			slog.Int64("post_id", event.PostID),
			slog.Time("timestamp", event.Timestamp),
		)

		hub.Broadcast(data)
	}

	handlePollVoted := func(data []byte) {
		var event core_events.PollVotedEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Warn("notification: failed to unmarshal poll.voted", slog.Any("error", err))
			return
		}
		log.Info("notification: poll voted",
			slog.Int64("poll_id", event.PollID),
			slog.Int64("option_id", event.OptionID),
			slog.Int64("user_id", event.UserID),
			slog.Time("timestamp", event.Timestamp),
		)

		hub.Broadcast(data)
	}

	subPostCreated, err := natsClient.Subscribe(core_events.SubjectPostCreated, handlePostCreated)
	if err != nil {
		log.Error("failed to subscribe", slog.String("subject", core_events.SubjectPostCreated), slog.Any("error", err))
		os.Exit(1)
	}
	defer subPostCreated.Unsubscribe()

	subPostDeleted, err := natsClient.Subscribe(core_events.SubjectPostDeleted, handlePostDeleted)
	if err != nil {
		log.Error("failed to subscribe", slog.String("subject", core_events.SubjectPostDeleted), slog.Any("error", err))
		os.Exit(1)
	}
	defer subPostDeleted.Unsubscribe()

	subPollVoted, err := natsClient.Subscribe(core_events.SubjectPollVoted, handlePollVoted)
	if err != nil {
		log.Error("failed to subscribe", slog.String("subject", core_events.SubjectPollVoted), slog.Any("error", err))
		os.Exit(1)
	}
	defer subPollVoted.Unsubscribe()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.ServeWS(hub, log))

	wsServer := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	go func() {
		log.Info("ws server listening", slog.String("addr", wsServer.Addr))
		if err := wsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("ws server failed", slog.Any("error", err))
			stop()
		}
	}()

	log.Info(">>> notification service STARTED")

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := wsServer.Shutdown(shutdownCtx); err != nil {
		log.Error("ws server shutdown failed", slog.Any("error", err))
	}

	log.Info(">>> notification service STOPPED")
}
