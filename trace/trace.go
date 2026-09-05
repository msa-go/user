package trace

import (
	// golang package
	"context"
	"fmt"
	"user/infrastructure/log"

	// external package
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer 는 OTLP HTTP exporter 로 트레이스를 내보내는 TracerProvider 를
// 설정하고, 종료 시 호출할 shutdown 함수를 반환한다.
func InitTracer(serviceName, endpoint string) (func(), error) {
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("init otlp exporter: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(provider)

	return func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			log.Logger.Errorf("error shutting down tracer: %v", err)
		}
	}, nil
}
