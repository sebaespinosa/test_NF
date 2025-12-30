package observability

import (
	"context"
	"fmt"
	"io"

	"github.com/sebaespinosa/test_NF/config"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// InitJaeger initializes Jaeger tracing using the Jaeger client library
func InitJaeger(ctx context.Context, cfg *config.JaegerConfig, serviceCfg *config.ServiceConfig) (io.Closer, error) {
	jCfg := jaegercfg.Configuration{
		ServiceName: serviceCfg.Name,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  cfg.SamplerType,
			Param: cfg.SamplerParam,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: fmt.Sprintf("%s:%d", cfg.AgentHost, cfg.AgentPort),
		},
	}

	closer, err := jCfg.InitGlobalTracer(serviceCfg.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize jaeger tracer: %w", err)
	}

	return closer, nil
}
