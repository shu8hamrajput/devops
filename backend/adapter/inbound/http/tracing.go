package http

import (
	"context"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// InitTracing initializes OpenTelemetry tracing with Jaeger exporter
func InitTracing(serviceName, jaegerEndpoint string) (*tracesdk.TracerProvider, error) {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	// Create resource with service name
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// TracingMiddleware returns a mux middleware for tracing HTTP requests
func TracingMiddleware() mux.MiddlewareFunc {
	return otelmux.Middleware("splitwise-backend")
}

