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

// Logger wraps slog.Logger and holds the log file reference
// so it can be properly closed on shutdown.
type Logger struct {
	*slog.Logger
	file *os.File
}

// contextKey is unexported to avoid collisions with other packages.
type contextKey struct{}

// FromContext extracts the Logger from context.
// Panics if logger was not injected — this is intentional:
// a missing logger in context is a programming error, not a runtime error.
func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(contextKey{}).(*Logger)
	if !ok {
		panic("logger: not found in context")
	}
	return log
}

// WithContext returns a new context with the Logger embedded.
func WithContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

// New creates a Logger that writes to both stdout and a timestamped log file.
//
// stdout — text (human-readable) in local/dev, JSON in prod
// file   — always JSON, for log aggregators (Loki, ELK, etc.)
//
// Call Close() on shutdown to close the file.
func New(cfg Config) (*Logger, error) {
	// --- parse log level ---
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("logger.New: unmarshal level %q: %w", cfg.Level, err)
	}

	// --- create log directory ---
	if err := os.MkdirAll(cfg.Folder, 0755); err != nil {
		return nil, fmt.Errorf("logger.New: create log folder: %w", err)
	}

	// --- open log file (one file per process start) ---
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000Z")
	logFilePath := filepath.Join(cfg.Folder, fmt.Sprintf("%s.log", timestamp))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("logger.New: open log file: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // adds file:line to every log entry
	}

	// Write to both stdout and file simultaneously
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// stdout gets text in local, JSON everywhere else
	var handler slog.Handler
	if cfg.Level == "debug" {
		// Text for local development — easier to read in terminal
		handler = slog.NewTextHandler(multiWriter, opts)
	} else {
		// JSON for dev/prod — works with log aggregators
		handler = slog.NewJSONHandler(multiWriter, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		file:   logFile,
	}, nil
}

// With returns a child logger with additional fields pre-attached.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		file:   l.file,
	}
}

// Close closes the log file.
// Must be called on application shutdown: defer logger.Close()
func (l *Logger) Close() {
	if err := l.file.Close(); err != nil {
		fmt.Println("logger: failed to close log file:", err)
	}
}
