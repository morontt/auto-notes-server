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
