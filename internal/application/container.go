package application

import (
	"log"
	"log/slog"
)

type Container struct {
	ErrorLog *log.Logger
	InfoLog  *slog.Logger
}
