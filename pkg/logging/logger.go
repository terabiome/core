package logging

import (
	"log/slog"
	"os"
	"strings"
)

type Slogger struct {
	*slog.Logger
}

var globalSlogger *Slogger

func NewSlogger(cfg Config, attributes map[string]any) *Slogger {
	var logLevel slog.Level

	switch strings.ToLower(cfg.Level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slogger := slog.New(handler)
	for k, v := range attributes {
		slogger = slogger.With(
			slog.Any(k, v),
		)
	}

	return &Slogger{
		Logger: slogger,
	}
}

func InitGlobalSlogger(cfg Config) {
	globalSlogger = NewSlogger(cfg, map[string]any{
		"scope": "global",
	})
}

func GetGlobalSlogger() *Slogger {
	return globalSlogger
}
