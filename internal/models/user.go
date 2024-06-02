package models

import "time"

type User struct {
	ID           uint
	Username     string
	PasswordHash string
	Salt         string
	CreatedAt    time.Time
}
