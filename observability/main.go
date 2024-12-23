package observability

import (
	"context"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// RegisterHooks registers the lifecycle hooks for telemetry providers
func RegisterHooks(lc fx.Lifecycle, providers *telemetryProviders, logging *otelzap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			err := providers.lp.ForceFlush(ctx)
			if err != nil {
				logging.Logger.Error("failed to flush log provider", zap.Error(err))
			}
			err = providers.tp.ForceFlush(ctx)
			if err != nil {
				logging.Logger.Error("failed to flush trace provider", zap.Error(err))
			}
			providers.cleanup()
			if err := providers.tp.Shutdown(ctx); err != nil {
				logging.Logger.Error("failed to shutdown trace provider", zap.Error(err))
			}
			if err := providers.lp.Shutdown(ctx); err != nil {
				logging.Logger.Error("failed to shutdown log provider", zap.Error(err))
			}
			return nil
		},
	})
}

var Module = fx.Module("observability",
	fx.Provide(
		NewLogging,
		NewTelemetryProviders,
	),
	fx.Invoke(RegisterHooks),
)
