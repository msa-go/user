package log

import (
	// golang package
	"context"

	// external package
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

var Logger *logrus.Logger

// SetupLogger setup logger
func SetupLogger() {
	log := logrus.New()

	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	log.Info("Logger initialized using logrus!")
	Logger = log
}

// LogWithTrace log with trace
func LogWithTrace(ctx context.Context) *logrus.Entry {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	return logrus.WithField("trace_id", traceID)
}
