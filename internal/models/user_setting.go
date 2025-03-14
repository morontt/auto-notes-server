package models

import (
	"database/sql"
	"time"
)

type UserSetting struct {
	ID         uint
	CarID      sql.NullInt32
	CarBrand   sql.NullString
	CarModel   sql.NullString
	CurrencyID sql.NullInt32
	CreatedAt  time.Time
	UpdatedAt  sql.NullTime
}
