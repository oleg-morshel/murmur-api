package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	*slog.Logger
	file *os.File
}

// contextKey is unexported to avoid collisions with other packages.
type contextKey struct{}

func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(contextKey{}).(*Logger)
	if !ok {
		panic("logger: not found in context")
	}
	return log
}

func WithContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

func New(cfg Config) (*Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("logger.New: unmarshal level %q: %w", cfg.Level, err)
	}

	if err := os.MkdirAll(cfg.Folder, 0755); err != nil {
		return nil, fmt.Errorf("logger.New: create log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000Z")
	logFilePath := filepath.Join(cfg.Folder, fmt.Sprintf("%s.log", timestamp))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("logger.New: open log file: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	var handler slog.Handler
	if cfg.Level == "debug" {
		handler = slog.NewTextHandler(multiWriter, opts)
	} else {
		handler = slog.NewJSONHandler(multiWriter, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		file:   logFile,
	}, nil
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		file:   l.file,
	}
}

func (l *Logger) Close() {
	if err := l.file.Close(); err != nil {
		fmt.Println("logger: failed to close log file:", err)
	}
}
