package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider is the main tracing provider for the application
type TracerProvider struct {
	tracer trace.Tracer
}

// InitTracer initializes OpenTelemetry tracing with an OTLP exporter.
func InitTracer(serviceName, otlpEndpoint string) (*TracerProvider, error) {
	// Set up OTLP HTTP exporter
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(otlpEndpoint), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Set up trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewSchemaless(
			attribute.String("service.name", serviceName),
		)),
	)

	// Set global Tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &TracerProvider{tracer: tp.Tracer(serviceName)}, nil
}

// StartSpan starts a new span for tracing.
func (tp *TracerProvider) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, spanName, opts...)
}

// EndSpan marks the end of a span and logs any errors
func (tp *TracerProvider) EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "success")
	}
	span.End()
}

// ShutdownTracer flushes any pending spans and stops the tracer provider.
func (tp *TracerProvider) ShutdownTracer(ctx context.Context) error {
	if tpProvider, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		return tpProvider.Shutdown(ctx)
	}
	return nil
}
