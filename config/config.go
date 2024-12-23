package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

type Config struct {
	Environment             string
	OtelEndpoint            string
	ServiceName             string
	LogLevel                string
	BatchSize               int
	ExportTimeout           string
	OtelExporterOtlpHeaders string
}

func NewConfig() (*Config, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// Get required environment variables
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return nil, fmt.Errorf("ENVIRONMENT environment variable is not set")
	}

	// Get OpenTelemetry configuration with defaults
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "http://localhost:4318" // Default OTLP HTTP endpoint
	}

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "unknown-service"
	}

	headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if headers == "" {
		headers = ""
	}

	// Logging configuration with defaults
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		if env == "production" {
			logLevel = "info"
		} else {
			logLevel = "debug"
		}
	}

	// Parse batch size with default
	batchSize := 512 // default batch size
	if batchSizeEnv := os.Getenv("OTEL_BATCH_SIZE"); batchSizeEnv != "" {
		if _, err := fmt.Sscanf(batchSizeEnv, "%d", &batchSize); err != nil {
			return nil, fmt.Errorf("invalid OTEL_BATCH_SIZE: %w", err)
		}
	}

	// Export timeout with default
	exportTimeout := os.Getenv("OTEL_EXPORT_TIMEOUT")
	if exportTimeout == "" {
		exportTimeout = "5s"
	}

	return &Config{
		Environment:   env,
		OtelEndpoint:  otelEndpoint,
		ServiceName:   serviceName,
		LogLevel:      logLevel,
		BatchSize:     batchSize,
		ExportTimeout: exportTimeout,
	}, nil
}

var Module = fx.Module("config", fx.Provide(NewConfig))
