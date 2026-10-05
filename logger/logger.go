package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func Setup(dev bool, level string) {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	if dev {
		opts := &tint.Options{
			Level:     slogLevel,
			AddSource: true,
			ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
				if attr.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return attr
			},
		}
		slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, opts)))
		return
	}

	opts := &slog.HandlerOptions{Level: slogLevel, AddSource: true}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, opts)))
}
