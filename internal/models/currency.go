package models

import "time"

type Currency struct {
	ID        uint
	Name      string
	Code      string
	Default   bool
	CreatedAt time.Time
}
