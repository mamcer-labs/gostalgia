package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceHandler wraps a slog.Handler and adds trace_id/span_id attributes
// from the active OTel span in ctx, so logs can be correlated with traces
// in Grafana (Tempo -> Loki) without touching every call site.
type TraceHandler struct {
	slog.Handler
}

func NewTraceHandler(h slog.Handler) *TraceHandler {
	return &TraceHandler{Handler: h}
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.TraceID().String()),
			slog.String("span_id", span.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
}
