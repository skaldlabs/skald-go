package skald

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type clientConfig struct {
	tracerProvider trace.TracerProvider
	httpClient     *http.Client
}

// Option configures the Skald client.
type Option func(*clientConfig)

// WithTracing enables OpenTelemetry tracing using the global TracerProvider.
func WithTracing() Option {
	return func(cfg *clientConfig) {
		cfg.tracerProvider = otel.GetTracerProvider()
	}
}

// WithTracerProvider enables OpenTelemetry tracing with a specific TracerProvider.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(cfg *clientConfig) {
		cfg.tracerProvider = tp
	}
}

// WithHTTPClient sets a custom HTTP client for the Skald client.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = client
	}
}
