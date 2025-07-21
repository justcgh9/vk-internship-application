package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
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

func FromContext(ctx context.Context) *slog.Logger {
	if Log == nil {
		Init(slog.LevelInfo)
	}
	span := trace.SpanFromContext(ctx)
	if span == nil || !span.SpanContext().HasTraceID() {
		return Log
	}
	return Log.With("trace_id", span.SpanContext().TraceID().String())
}
