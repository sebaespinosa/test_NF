package observability

import (
	"context"
	"fmt"
	"net"

	"github.com/sebaespinosa/test_NF/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitJaeger initializes OpenTelemetry with OTLP exporter for Jaeger
func InitJaeger(ctx context.Context, cfg *config.JaegerConfig, serviceCfg *config.ServiceConfig) (func(context.Context) error, error) {
	// Create OTLP gRPC exporter
	host := cfg.AgentHost
	if host == "localhost" {
		// Force IPv4 to avoid ::1 refusals when container only listens on 0.0.0.0
		host = "127.0.0.1"
	}
	endpoint := net.JoinHostPort(host, "4317")

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceCfg.Name),
			semconv.ServiceVersion(serviceCfg.Version),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Return shutdown function
	return tp.Shutdown, nil
}
