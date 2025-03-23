package examples

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/h2co32/gollama/pkg/observability"
)

// ObservabilityBasicExample demonstrates basic usage of the observability package.
func ObservabilityBasicExample() {
	// Initialize a tracer provider
	// In a real application, you would use a real collector endpoint
	tp, err := observability.NewTracerProvider("example-service", "localhost:4318")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create a root span
	ctx, span := tp.StartSpan(context.Background(), "main-operation")
	defer span.End()

	// Add attributes to the span
	observability.AddSpanAttributes(ctx,
		attribute.String("environment", "development"),
		attribute.Int("user_id", 123),
	)

	// Add an event to the span
	observability.AddSpanEvent(ctx, "operation-started", 
		attribute.String("correlation_id", "abc-123"),
	)

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Create a child span for a sub-operation
	err = observability.WithSpan(ctx, "sub-operation", func(ctx context.Context) error {
		// Simulate some work in the sub-operation
		time.Sleep(50 * time.Millisecond)
		
		// Add an event to the child span
		observability.AddSpanEvent(ctx, "sub-operation-event")
		
		return nil
	})

	if err != nil {
		observability.AddSpanError(ctx, err)
		observability.SetSpanStatus(ctx, codes.Error, "Operation failed")
	} else {
		observability.SetSpanStatus(ctx, codes.Ok, "Operation succeeded")
	}

	fmt.Println("Tracing example completed")
}

// ObservabilityHTTPExample demonstrates tracing HTTP requests.
func ObservabilityHTTPExample() {
	// Initialize a tracer provider
	tp, err := observability.NewTracerProvider("http-service", "localhost:4318")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create a traced HTTP handler
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		// Extract the context from the request (in a real app, you would use
		// OpenTelemetry's HTTP middleware to propagate the context)
		ctx := r.Context()
		
		// Create a span for this handler
		ctx, span := tp.StartSpan(ctx, "hello-handler")
		defer span.End()
		
		// Add request information to the span
		observability.AddSpanAttributes(ctx,
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.Path),
			attribute.String("http.user_agent", r.UserAgent()),
		)
		
		// Simulate a database query
		err := observability.WithSpanTimed(ctx, "database-query", func(ctx context.Context) error {
			// Simulate database work
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		
		if err != nil {
			observability.AddSpanError(ctx, err)
			observability.SetSpanStatus(ctx, codes.Error, "Database query failed")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		
		// Write response
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
		
		observability.SetSpanStatus(ctx, codes.Ok, "Request processed successfully")
	})

	// In a real application, you would start the server:
	// log.Fatal(http.ListenAndServe(":8080", nil))
	
	fmt.Println("HTTP tracing example setup completed")
	fmt.Println("In a real application, the server would be running at http://localhost:8080/hello")
}

// ObservabilityErrorHandlingExample demonstrates error handling with tracing.
func ObservabilityErrorHandlingExample() {
	// Initialize a tracer provider with custom options
	options := observability.TracerOptions{
		SamplingRatio:     0.5, // Sample 50% of traces
		ServiceNamespace:  "examples",
		ServiceVersion:    "1.0.0",
		AdditionalAttributes: []attribute.KeyValue{
			attribute.String("deployment.environment", "development"),
		},
	}
	
	tp, err := observability.NewTracerProviderWithOptions("error-service", "localhost:4318", options)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Create a root span
	ctx, span := tp.StartSpan(context.Background(), "process-data")
	defer span.End()

	// Simulate a series of operations with an error
	err = observability.WithSpan(ctx, "fetch-data", func(ctx context.Context) error {
		// Simulate successful data fetching
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	
	if err != nil {
		observability.AddSpanError(ctx, err)
		observability.SetSpanStatus(ctx, codes.Error, "Failed to fetch data")
		fmt.Println("Error fetching data:", err)
		return
	}
	
	err = observability.WithSpan(ctx, "transform-data", func(ctx context.Context) error {
		// Simulate an error during data transformation
		time.Sleep(30 * time.Millisecond)
		return fmt.Errorf("data transformation failed: invalid format")
	})
	
	if err != nil {
		// Record the error in the span
		observability.AddSpanError(ctx, err)
		observability.SetSpanStatus(ctx, codes.Error, "Failed to transform data")
		fmt.Println("Error transforming data:", err)
		
		// In a real application, you might perform error recovery or fallback here
		err = observability.WithSpan(ctx, "error-recovery", func(ctx context.Context) error {
			// Simulate error recovery logic
			time.Sleep(20 * time.Millisecond)
			return nil
		})
		
		if err != nil {
			observability.AddSpanError(ctx, err)
			fmt.Println("Error recovery failed:", err)
		} else {
			fmt.Println("Error recovery succeeded")
		}
		
		return
	}
	
	// This won't execute due to the error above
	err = observability.WithSpan(ctx, "save-data", func(ctx context.Context) error {
		// Simulate saving data
		time.Sleep(40 * time.Millisecond)
		return nil
	})
	
	if err != nil {
		observability.AddSpanError(ctx, err)
		observability.SetSpanStatus(ctx, codes.Error, "Failed to save data")
		fmt.Println("Error saving data:", err)
		return
	}
	
	observability.SetSpanStatus(ctx, codes.Ok, "Data processed successfully")
	fmt.Println("Data processing completed successfully")
}
