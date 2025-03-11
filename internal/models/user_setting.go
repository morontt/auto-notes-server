package models

import (
	"database/sql"
	"time"
)

type UserSetting struct {
	ID         uint
	CarID      sql.NullInt32
	CurrencyID sql.NullInt32
	CreatedAt  time.Time
	UpdatedAt  sql.NullTime
}
