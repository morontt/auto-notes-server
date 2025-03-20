package models

import (
	"database/sql"
	"time"
)

type UserSetting struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt sql.NullTime

	CarID    sql.NullInt32
	CarBrand sql.NullString
	CarModel sql.NullString

	CurrencyID        sql.NullInt32
	CurrencyName      sql.NullString
	CurrencyCode      sql.NullString
	CurrencyCreatedAt sql.NullTime
}
