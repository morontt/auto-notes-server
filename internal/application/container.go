package application

import (
	"database/sql"
	"log/slog"
)

const (
	CtxKeyUser      = "user_ctx_key"
	CtxKeyRequestID = "req_id_ctx_key"
)

type Container struct {
	DB     *sql.DB
	logger *slog.Logger
}

func (c *Container) SetupDatabase() error {
	db, err := getDBConnection(c.logger)
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
	c.logger.Info("The database connection is closed")

	return nil
}
