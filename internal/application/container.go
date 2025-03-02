package application

import (
	"database/sql"
	"log"
	"log/slog"
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
