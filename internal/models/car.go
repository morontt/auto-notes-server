package models

import (
	"database/sql"
	"time"
)

type Car struct {
	ID        uint
	Brand     string
	Model     string
	Year      sql.NullInt32
	Vin       sql.NullString
	UserID    uint
	Default   bool
	CreatedAt time.Time
	UpdatedAt sql.NullTime
}
