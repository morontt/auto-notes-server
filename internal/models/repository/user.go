package repository

import (
	"database/sql"
	"errors"

	"xelbot.com/auto-notes/server/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func (ur *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.password,
			u.password_salt,
			u.created_at
		FROM users AS u
		WHERE (u.username = ?)`

	user := models.User{}

	err := ur.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Salt,
		&user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	return &user, nil
}
