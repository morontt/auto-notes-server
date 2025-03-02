package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"runtime/debug"
)

const (
	CtxKeyUser      = "user_ctx_key"
	CtxKeyRequestID = "req_id_ctx_key"
)

type Container struct {
	ErrorLog *log.Logger
	InfoLog  *slog.Logger
	DB       *sql.DB
}

func (c *Container) Info(msg string, ctx context.Context, args ...any) {
	c.InfoLog.Info(msg, logContext(ctx, args...)...)
}

func (c *Container) Debug(msg string, ctx context.Context, args ...any) {
	c.InfoLog.Debug(msg, logContext(ctx, args...)...)
}

func (c *Container) Warn(msg string, ctx context.Context, args ...any) {
	c.InfoLog.Warn(msg, logContext(ctx, args...)...)
}

func (c *Container) ServerError(err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	c.ErrorLog.Println(trace)
}

func (c *Container) SetupDatabase() error {
	db, err := getDBConnection(c.InfoLog)
	if err != nil {
		return err
	}

	c.DB = db

	return nil
}

func (c *Container) Stop() error {
	err := c.DB.Close()
	if err != nil {
		return err
	}
	c.InfoLog.Info("The database connection is closed")

	return nil
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
