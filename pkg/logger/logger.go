package logger

import (
	"log/slog"
	"os"
)

var (
	Log *slog.Logger
)

func Init(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	Log = slog.New(handler)
}
