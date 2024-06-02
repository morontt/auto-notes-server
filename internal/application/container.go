package application

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"runtime/debug"
)

type Container struct {
	ErrorLog *log.Logger
	InfoLog  *slog.Logger
	DB       *sql.DB
}

func (c *Container) Info(msg string, args ...any) {
	c.InfoLog.Info(msg, args...)
}

func (c *Container) Debug(msg string, args ...any) {
	c.InfoLog.Debug(msg, args...)
}

func (c *Container) ServerError(err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	c.ErrorLog.Println(trace)
}
