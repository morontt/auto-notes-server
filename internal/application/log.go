package application

import (
	"context"
	"fmt"
	"runtime/debug"
)

func (c *Container) Info(msg string, ctx context.Context, args ...any) {
	c.InfoLog.Info(msg, logContext(ctx, args...)...)
}

func (c *Container) Debug(msg string, ctx context.Context, args ...any) {
	c.InfoLog.Debug(msg, logContext(ctx, args...)...)
}

func (c *Container) Warn(msg string, ctx context.Context, args ...any) {
	c.InfoLog.Warn(msg, logContext(ctx, args...)...)
}

func (c *Container) Error(msg string, ctx context.Context, args ...any) {
	c.InfoLog.Error(msg, logContext(ctx, args...)...)
}

func (c *Container) ServerError(err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	c.ErrorLog.Println(trace)
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
