package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/fleet-platform/shared/pkg/logger"
)

// InitProvider sets up the OpenTelemetry trace provider globally
func InitProvider(ctx context.Context, log logger.Logger, serviceName string, otlpEndpoint string) (func(context.Context) error, error) {
	if otlpEndpoint == "" {
		log.Warn("OTLP endpoint is empty, tracing is disabled")
		return func(context.Context) error { return nil }, nil
	}

	// Create OTLP exporter (gRPC)
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(otlpEndpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Register the trace provider
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider.Shutdown, nil
}

// WrapHTTPClient wraps an existing HTTP client with OpenTelemetry tracing
func WrapHTTPClient(client *http.Client) *http.Client {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}
	client.Transport = otelhttp.NewTransport(client.Transport)
	return client
}

// WrapHTTPHandler wraps an http.Handler with OpenTelemetry tracing
func WrapHTTPHandler(h http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(h, operation)
}
