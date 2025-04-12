package application

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
)

func (c *Container) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

func (c *Container) Info(msg string, ctx context.Context, args ...any) {
	c.logger.Info(msg, logContext(ctx, args...)...)
}

func (c *Container) Debug(msg string, ctx context.Context, args ...any) {
	c.logger.Debug(msg, logContext(ctx, args...)...)
}

func (c *Container) Warn(msg string, ctx context.Context, args ...any) {
	c.logger.Warn(msg, logContext(ctx, args...)...)
}

func (c *Container) Error(msg string, ctx context.Context, args ...any) {
	c.logger.Error(msg, logContext(ctx, args...)...)
}

func (c *Container) ServerError(ctx context.Context, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	c.logger.Error("server internal error", logContext(ctx, "trace", trace)...)
}

func logContext(ctx context.Context, args ...any) []any {
	var additional []any
	if reqID, ok := ctx.Value(CtxKeyRequestID).(string); ok {
		additional = append(additional, "request_id", reqID)
	}

	if len(additional) > 0 {
		args = append(args, additional...)
	}

	return args
}
