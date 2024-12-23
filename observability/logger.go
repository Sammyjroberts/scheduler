package observability

import (
	"context"
	"fmt"
	"time"
	"turionspace/nei-mission-planner/scheduler/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type telemetryProviders struct {
	tp      *sdktrace.TracerProvider
	lp      *sdklog.LoggerProvider
	cleanup func()
}

// NewLogging creates a new logging instance without any fx lifecycle bindings
func NewLogging(cfg *config.Config) (*otelzap.Logger, error) {
	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		return nil, err
	}

	// Wrap with OpenTelemetry
	otelLogger := otelzap.New(logger)

	return otelLogger, nil
}

// NewTelemetryProviders initializes OpenTelemetry providers
func NewTelemetryProviders(cfg *config.Config) (*telemetryProviders, error) {
	spew.Dump(cfg)
	cleanup, tp, lp, err := initOpenTelemetry(cfg)
	if err != nil {
		return nil, err
	}

	return &telemetryProviders{
		tp:      tp,
		lp:      lp,
		cleanup: cleanup,
	}, nil
}

func initOpenTelemetry(cfg *config.Config) (cleanup func(), tp *sdktrace.TracerProvider, lp *sdklog.LoggerProvider, err error) {
	ctx := context.Background()
	// Test connection before creating exporter
	conn, err := grpc.Dial(
		cfg.OtelEndpoint,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		fmt.Printf("Warning: Failed to connect to OTLP endpoint: %v\n", err)
	} else {
		conn.Close()
		fmt.Printf("Successfully connected to OTLP endpoint\n")
	}
	// Initialize OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(), // TODO: make secure for production
		otlptracegrpc.WithEndpoint(cfg.OtelEndpoint),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	// Initialize OTLP log exporter
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(cfg.OtelEndpoint),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create trace provider
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithMaxExportBatchSize(cfg.BatchSize),
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Create log provider
	lp = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(logExporter),
		),
	)

	// Set global providers
	otel.SetTracerProvider(tp)
	global.SetLoggerProvider(lp)

	return func() {}, tp, lp, nil
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var config zap.Config

	// Set development or production config based on environment
	if cfg.Environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level from config
	level, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Build logger
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("service", cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

var LoggerModule = fx.Provide("logger", NewLogging)
