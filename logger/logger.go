// Package logger configures application logging.
package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"

	"afastchat.com/env"
	"github.com/lmittmann/tint"
)

func Setup() {
	w := os.Stderr

	if env.Env.Dev {
		opts := &tint.Options{
			// Level:     slog.LevelDebug,
			Level:     slog.LevelInfo,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return a
			},
			TimeFormat: "",
			NoColor:    false,
		}
		slog.SetDefault(slog.New(tint.NewHandler(w, opts)))
		return
	}

	opts := &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		AddSource:   true,
		ReplaceAttr: nil,
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(w, opts)))
}

func PrettyJSON(s string) string {
	var obj any
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return s
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return s
	}
	return buf.String()
}
