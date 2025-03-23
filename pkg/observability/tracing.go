// Package observability provides tools for distributed tracing and metrics collection.
//
// This package integrates with OpenTelemetry to provide distributed tracing capabilities,
// allowing you to track requests across service boundaries and understand the flow of
// operations in your application.
//
// Example usage:
//
//	// Initialize a tracer provider
//	tp, err := observability.NewTracerProvider("my-service", "http://localhost:4318")
//	if err != nil {
//		log.Fatalf("Failed to initialize tracer: %v", err)
//	}
//	defer tp.Shutdown(context.Background())
//
//	// Create a span
//	ctx, span := tp.Tracer().Start(context.Background(), "my-operation")
//	defer span.End()
//
//	// Add attributes to the span
//	span.SetAttributes(attribute.String("key", "value"))
//
//	// Create a child span
//	childCtx, childSpan := tp.Tracer().Start(ctx, "child-operation")
//	defer childSpan.End()
package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Version represents the current package version following semantic versioning.
const Version = "1.0.0"

// TracerProvider wraps the OpenTelemetry TracerProvider with additional functionality.
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// TracerOptions configures the TracerProvider.
type TracerOptions struct {
	// SamplingRatio sets the sampling ratio for traces (0.0 to 1.0).
	// Default: 1.0 (sample all traces)
	SamplingRatio float64

	// ServiceNamespace is an optional namespace for the service.
	ServiceNamespace string

	// ServiceVersion is the version of the service.
	// Default: "unknown"
	ServiceVersion string

	// AdditionalAttributes are additional resource attributes to include with all spans.
	AdditionalAttributes []attribute.KeyValue
}

// DefaultTracerOptions returns the default tracer options.
func DefaultTracerOptions() TracerOptions {
	return TracerOptions{
		SamplingRatio:  1.0,
		ServiceVersion: "unknown",
	}
}

// NewTracerProvider creates a new TracerProvider with the specified service name and endpoint.
// The endpoint should be the URL of the OpenTelemetry collector, e.g., "http://localhost:4318".
func NewTracerProvider(serviceName, endpoint string) (*TracerProvider, error) {
	return NewTracerProviderWithOptions(serviceName, endpoint, DefaultTracerOptions())
}

// NewTracerProviderWithOptions creates a new TracerProvider with custom options.
func NewTracerProviderWithOptions(serviceName, endpoint string, options TracerOptions) (*TracerProvider, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	// Configure the OTLP exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // For development; use TLS in production
	)

	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create resource attributes
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(options.ServiceVersion),
	}

	if options.ServiceNamespace != "" {
		attrs = append(attrs, semconv.ServiceNamespaceKey.String(options.ServiceNamespace))
	}

	// Add additional attributes
	attrs = append(attrs, options.AdditionalAttributes...)

	// Create a resource
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)

	// Configure the trace provider
	samplingRatio := options.SamplingRatio
	if samplingRatio <= 0 || samplingRatio > 1 {
		samplingRatio = 1.0
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(samplingRatio)),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set the global trace provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create a tracer
	tracer := tp.Tracer(serviceName, trace.WithInstrumentationVersion(Version))

	return &TracerProvider{
		provider: tp,
		tracer:   tracer,
	}, nil
}

// Tracer returns the tracer instance.
func (tp *TracerProvider) Tracer() trace.Tracer {
	return tp.tracer
}

// Shutdown shuts down the tracer provider, flushing any remaining spans.
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	return tp.provider.Shutdown(ctx)
}

// StartSpan starts a new span with the given name.
// It returns the context with the span and the span itself.
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// AddSpanEvent adds an event to the current span in the context.
func AddSpanEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// AddSpanError adds an error to the current span in the context.
func AddSpanError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

// SetSpanStatus sets the status of the current span in the context.
func SetSpanStatus(ctx context.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(code, description)
}

// AddSpanAttributes adds attributes to the current span in the context.
func AddSpanAttributes(ctx context.Context, attributes ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attributes...)
}

// WithSpan wraps a function with a new span.
// It creates a span, executes the function, and ends the span when the function returns.
func WithSpan(ctx context.Context, name string, fn func(context.Context) error) error {
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
	}

	return err
}

// WithSpanTimed wraps a function with a new span and records the execution time.
func WithSpanTimed(ctx context.Context, name string, fn func(context.Context) error) error {
	tracer := otel.Tracer("")
	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	elapsed := time.Since(start)

	span.SetAttributes(attribute.Int64("duration_ms", elapsed.Milliseconds()))
	if err != nil {
		span.RecordError(err)
	}

	return err
}
