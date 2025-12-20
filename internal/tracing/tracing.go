// Package tracing provides OpenTelemetry distributed tracing support for the application
package tracing

import (
	"context"
	"subsnotifpro-go/internal/pkg/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingConfig holds configuration for tracing
type TracingConfig struct {
	ServiceName      string
	ServiceVersion   string
	Environment      string
	OTLPEndpoint     string
	SampleRate       float64
	Enabled          bool
}

// TracerProvider wraps OpenTelemetry tracer provider with additional functionality
type TracerProvider struct {
	provider   *trace.TracerProvider
	tracer     oteltrace.Tracer
	config     TracingConfig
	shutdown   func(context.Context) error
}

// NewTracerProvider creates and configures a new OpenTelemetry tracer provider
func NewTracerProvider(config TracingConfig) (*TracerProvider, error) {
	if !config.Enabled {
		logger.Log.Info("🔍 Tracing disabled by configuration")
		return &TracerProvider{
			provider: trace.NewTracerProvider(trace.WithSampler(trace.NeverSample())),
			tracer:   otel.Tracer(config.ServiceName),
			config:   config,
			shutdown: func(context.Context) error { return nil },
		}, nil
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			resource.Default().SchemaURL(),
			attribute.String("service.name", config.ServiceName),
			attribute.String("service.version", config.ServiceVersion),
			attribute.String("deployment.environment", config.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider with OTLP exporter
	var tp *trace.TracerProvider
	var shutdown func(context.Context) error

	tp, shutdown, err = createOTLPTracerProvider(res, config)
	if err != nil {
		logger.Log.Warnf("Failed to create OTLP tracer, using noop tracer: %v", err)
		tp = trace.NewTracerProvider(trace.WithSampler(trace.NeverSample()))
		shutdown = func(context.Context) error { return nil }
	}

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(config.ServiceName)

	logger.Log.Info("🔍 Tracing initialized with OTLP exporter")

	return &TracerProvider{
		provider: tp,
		tracer:   tracer,
		config:   config,
		shutdown: shutdown,
	}, nil
}

// createOTLPTracerProvider creates a tracer provider with OTLP exporter
func createOTLPTracerProvider(res *resource.Resource, config TracingConfig) (*trace.TracerProvider, func(context.Context) error, error) {
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // Use TLS in production
	)
	if err != nil {
		return nil, nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(config.SampleRate)),
	)

	return tp, tp.Shutdown, nil
}

// Tracer returns the underlying OpenTelemetry tracer
func (tp *TracerProvider) Tracer() oteltrace.Tracer {
	return tp.tracer
}

// StartSpan starts a new span with optional attributes
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes starts a new span with predefined attributes
func (tp *TracerProvider) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]interface{}) (context.Context, oteltrace.Span) {
	attributes := make([]attribute.KeyValue, 0, len(attrs))
	for k, v := range attrs {
		switch val := v.(type) {
		case string:
			attributes = append(attributes, attribute.String(k, val))
		case int:
			attributes = append(attributes, attribute.Int(k, val))
		case int64:
			attributes = append(attributes, attribute.Int64(k, val))
		case float64:
			attributes = append(attributes, attribute.Float64(k, val))
		case bool:
			attributes = append(attributes, attribute.Bool(k, val))
		default:
			attributes = append(attributes, attribute.String(k, "unknown"))
		}
	}

	return tp.tracer.Start(ctx, name, oteltrace.WithAttributes(attributes...))
}

// RecordError records an error on the span from context
func (tp *TracerProvider) RecordError(ctx context.Context, err error, description string) {
	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err, oteltrace.WithAttributes(
			attribute.String("error.description", description),
		))
		span.SetStatus(codes.Error, description)
	}
}

// AddEvent adds an event to the span from context
func (tp *TracerProvider) AddEvent(ctx context.Context, name string, attrs map[string]interface{}) {
	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		attributes := make([]attribute.KeyValue, 0, len(attrs))
		for k, v := range attrs {
			switch val := v.(type) {
			case string:
				attributes = append(attributes, attribute.String(k, val))
			case int:
				attributes = append(attributes, attribute.Int(k, val))
			case int64:
				attributes = append(attributes, attribute.Int64(k, val))
			case float64:
				attributes = append(attributes, attribute.Float64(k, val))
			case bool:
				attributes = append(attributes, attribute.Bool(k, val))
			}
		}
		span.AddEvent(name, oteltrace.WithAttributes(attributes...))
	}
}

// SetAttributes sets attributes on the span from context
func (tp *TracerProvider) SetAttributes(ctx context.Context, attrs map[string]interface{}) {
	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		for k, v := range attrs {
			switch val := v.(type) {
			case string:
				span.SetAttributes(attribute.String(k, val))
			case int:
				span.SetAttributes(attribute.Int(k, val))
			case int64:
				span.SetAttributes(attribute.Int64(k, val))
			case float64:
				span.SetAttributes(attribute.Float64(k, val))
			case bool:
				span.SetAttributes(attribute.Bool(k, val))
			}
		}
	}
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.shutdown != nil {
		return tp.shutdown(ctx)
	}
	return nil
}

// TraceHTTPRequest is a helper to trace HTTP requests
func (tp *TracerProvider) TraceHTTPRequest(ctx context.Context, method, url string) (context.Context, oteltrace.Span) {
	return tp.StartSpanWithAttributes(ctx, "http.request", map[string]interface{}{
		"http.method": method,
		"http.url":    url,
		"component":   "http.client",
	})
}

// TraceDBQuery is a helper to trace database queries
func (tp *TracerProvider) TraceDBQuery(ctx context.Context, operation, table string) (context.Context, oteltrace.Span) {
	return tp.StartSpanWithAttributes(ctx, "db.query", map[string]interface{}{
		"db.operation": operation,
		"db.table":     table,
		"component":    "database",
	})
}

// TraceMessageProcessing is a helper to trace message queue processing
func (tp *TracerProvider) TraceMessageProcessing(ctx context.Context, queueName, messageType string) (context.Context, oteltrace.Span) {
	return tp.StartSpanWithAttributes(ctx, "message.process", map[string]interface{}{
		"messaging.queue":        queueName,
		"messaging.message_type": messageType,
		"component":              "messaging",
	})
}

// TraceBusinessOperation is a helper to trace business operations
func (tp *TracerProvider) TraceBusinessOperation(ctx context.Context, operation string, attrs map[string]interface{}) (context.Context, oteltrace.Span) {
	if attrs == nil {
		attrs = make(map[string]interface{})
	}
	attrs["operation.type"] = "business"
	attrs["operation.name"] = operation
	
	return tp.StartSpanWithAttributes(ctx, operation, attrs)
}

// GetTraceID returns the trace ID from the current span context
func (tp *TracerProvider) GetTraceID(ctx context.Context) string {
	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the current span context
func (tp *TracerProvider) GetSpanID(ctx context.Context) string {
	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		return span.SpanContext().SpanID().String()
	}
	return ""
}
