package repository

import (
	"database/sql"
	"errors"

	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/utils/database"
)

type UserRepository struct {
	DB *database.DB
}

func (ur *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.password,
			u.created_at
		FROM users AS u
		WHERE (u.username = ?)`

	obj := models.User{}

	err := ur.DB.QueryRow(query, username).Scan(
		&obj.ID,
		&obj.Username,
		&obj.PasswordHash,
		&obj.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	return &obj, nil
}
