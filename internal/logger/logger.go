package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	defaultOnce   sync.Once
	defaultLevel  = slog.LevelInfo
	loggerKey     = "youtube-tracker/logger"
)

func Init(w io.Writer, level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl}))
	slog.SetDefault(logger)
	defaultLogger = logger
	defaultLevel = lvl

	return logger
}

func Default() *slog.Logger {
	defaultOnce.Do(func() {
		defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: defaultLevel}))
	})
	return defaultLogger
}

func WithContext(ctx context.Context) context.Context {
	return WithContextAttrs(Default(), ctx)
}

func WithContextAttrs(logger *slog.Logger, ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return Default()
}

func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

func Err(err error, msg string, args ...any) {
	Default().Error(msg, append([]any{"error", err}, args...)...)
}
