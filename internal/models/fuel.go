package models

import (
	"database/sql"
	"time"
)

type FillingStation struct {
	ID        uint
	Name      string
	CreatedAt time.Time
}

type Fuel struct {
	ID        uint
	Cost      Cost
	Value     int32
	Station   FillingStation
	Date      time.Time
	Distance  sql.NullInt32
	Car       Car
	CreatedAt time.Time
}
