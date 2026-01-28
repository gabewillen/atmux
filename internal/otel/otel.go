// Package otel implements OpenTelemetry scaffolding for the amux project
package otel

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Provider holds the OpenTelemetry trace and metric providers
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	serviceName    string
}

// NewProvider creates a new OpenTelemetry provider with the given service name
func NewProvider(serviceName string, resourceAttrs ...attribute.KeyValue) (*Provider, error) {
	// Create resource with service name and other attributes
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		append([]attribute.KeyValue{
			semconv.ServiceName(serviceName),
		}, resourceAttrs...)...,
	)

	// Create trace provider with stdout exporter (for now)
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Create metric provider with stdout exporter (for now)
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)

	// Set global providers
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &Provider{
		tracerProvider: tp,
		meterProvider:  mp,
		serviceName:    serviceName,
	}, nil
}

// Shutdown shuts down the OpenTelemetry providers
func (p *Provider) Shutdown(ctx context.Context) error {
	// Shutdown trace provider
	if err := p.tracerProvider.Shutdown(ctx); err != nil {
		return err
	}

	// Shutdown meter provider
	if err := p.meterProvider.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

// TracerProvider returns the trace provider
func (p *Provider) TracerProvider() *sdktrace.TracerProvider {
	return p.tracerProvider
}

// MeterProvider returns the metric provider
func (p *Provider) MeterProvider() *sdkmetric.MeterProvider {
	return p.meterProvider
}

// GetTracer returns a tracer with the given name
func (p *Provider) GetTracer(name string) trace.Tracer {
	return p.tracerProvider.Tracer(name)
}

// GetMeter returns a meter with the given name
func (p *Provider) GetMeter(name string) metric.Meter {
	return p.meterProvider.Meter(name)
}