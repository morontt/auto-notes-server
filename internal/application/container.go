package application

import (
	"log/slog"

	"xelbot.com/auto-notes/server/internal/utils/database"
)

const (
	CtxKeyUser      = "user_ctx_key"
	CtxKeyRequestID = "req_id_ctx_key"
)

type Container struct {
	DB     *database.DB
	logger *slog.Logger
}

func (c *Container) SetupDatabase() error {
	db, err := getDBConnection(c.logger)
	if err != nil {
		return err
	}

	c.DB = database.Wrap(db, c.logger)

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
