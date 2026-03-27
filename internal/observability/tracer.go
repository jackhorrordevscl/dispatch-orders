package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitTracer configura OpenTelemetry con un exporter a stdout (JSON)
// En producción se reemplazaría por OTLP hacia Jaeger, Datadog, etc.
func InitTracer(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
    // Exporter: escribe trazas como JSON en stdout
    exporter, err := stdouttrace.New(
        stdouttrace.WithPrettyPrint(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create trace exporter: %w", err)
    }

    // Resource: metadatos del servicio
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
            semconv.ServiceVersion("1.0.0"),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }

    // TracerProvider: orquesta el sampling y exportación
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(trace.AlwaysSample()),
    )

    // Registrar como provider global
    otel.SetTracerProvider(tp)

    return tp, nil
}