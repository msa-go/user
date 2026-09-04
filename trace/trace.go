package trace

import (
	// golang package
	"context"
	"log"

	// external package
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitTracer(serviceName string) (func(), error) {
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)

	if err != nil {
		log.Fatalf("Error Init Tracer %v", err)
	}

	trace := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(trace)
	return func() {
		if err := trace.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down tracer: %v", err)
		}
	}, nil
}
