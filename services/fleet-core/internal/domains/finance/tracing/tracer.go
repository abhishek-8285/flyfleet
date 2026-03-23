package tracing

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "flyfleet-finance"

var Tracer trace.Tracer

// InitTracer initializes the OpenTelemetry tracer for the finance domain.
// In production, export to Jaeger or Google Cloud Trace.
func InitTracer() (func(context.Context) error, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("service.domain", "finance"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	options := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	}

	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint != "" {
		exp, expErr := otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint(strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")),
			otlptracehttp.WithInsecure(),
		)
		if expErr == nil {
			options = append(options, sdktrace.WithBatcher(exp))
		}
	}

	tp := sdktrace.NewTracerProvider(options...)

	otel.SetTracerProvider(tp)
	Tracer = tp.Tracer(serviceName)

	return tp.Shutdown, nil
}

// StartSpan creates a new span for tracing a finance operation.
func StartSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	if Tracer == nil {
		Tracer = otel.Tracer(serviceName)
	}
	return Tracer.Start(ctx, operation, trace.WithAttributes(attrs...))
}
