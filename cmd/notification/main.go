package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/oleg-morshel/murmur-api/internal/config"
	core_events "github.com/oleg-morshel/murmur-api/internal/core/events"
	core_nats "github.com/oleg-morshel/murmur-api/internal/core/queue/nats"
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

	natsClient, err := core_nats.NewClient(ctx, core_nats.NewConfigMust(), log)
	if err != nil {
		log.Error("failed to init nats", slog.Any("error", err))
		os.Exit(1)
	}
	defer natsClient.Close()

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

	log.Info(">>> notification service STARTED")

	<-ctx.Done()

	log.Info(">>> notification service STOPPED")
}
