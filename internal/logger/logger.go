package logger

import (
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
