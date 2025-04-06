package repository

import (
	"database/sql"
	"errors"

	"xelbot.com/auto-notes/server/internal/models"
)

type CurrencyRepository struct {
	DB *sql.DB
}

func (cr *CurrencyRepository) GetCurrencies(userID uint) ([]*models.Currency, error) {
	query := `
		SELECT
			c.id,
			c.name,
			c.code,
			IF(s.id IS NULL, 0, 1) AS is_default,
			c.created_at
		FROM currencies AS c
		LEFT JOIN user_settings AS s ON (c.id = s.default_currency_id AND s.user_id = ?)
		ORDER BY c.name`

	rows, err := cr.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]*models.Currency, 0)

	for rows.Next() {
		obj := models.Currency{}
		err = rows.Scan(
			&obj.ID,
			&obj.Name,
			&obj.Code,
			&obj.Default,
			&obj.CreatedAt)

		if err != nil {
			return nil, err
		}

		items = append(items, &obj)
	}

	return items, nil
}

func (cr *CurrencyRepository) GetCurrencyByCode(code string) (*models.Currency, error) {
	query := `
		SELECT
			c.id,
			c.name,
			c.code,
			c.created_at
		FROM currencies AS c
		WHERE c.code = ?`

	obj := models.Currency{}

	err := cr.DB.QueryRow(query, code).Scan(
		&obj.ID,
		&obj.Name,
		&obj.Code,
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
