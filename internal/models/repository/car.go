package repository

import (
	"database/sql"
	"errors"

	"xelbot.com/auto-notes/server/internal/models"
)

type CarRepository struct {
	DB *sql.DB
}

func (cr *CarRepository) GetCarsByUser(userID uint) ([]*models.Car, error) {
	query := `
		SELECT
			c.id,
			c.brand_name,
			c.model_name,
			c.prod_year,
			c.vin,
			IF(c.id = s.default_car_id, 1, 0) AS is_default,
			c.created_at,
			c.updated_at
		FROM cars AS c
		LEFT JOIN user_settings AS s ON c.user_id = s.user_id
		WHERE c.user_id = ?
		ORDER BY c.id DESC`

	rows, err := cr.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	cars := make([]*models.Car, 0)

	for rows.Next() {
		car := models.Car{}
		err = rows.Scan(
			&car.ID,
			&car.Brand,
			&car.Model,
			&car.Year,
			&car.Vin,
			&car.Default,
			&car.CreatedAt,
			&car.UpdatedAt)

		if err != nil {
			return nil, err
		}

		cars = append(cars, &car)
	}

	return cars, nil
}

func (cr *CarRepository) Find(id uint) (*models.Car, error) {
	query := `
		SELECT
			c.id,
			c.user_id,
			c.created_at
		FROM cars AS c
		WHERE c.id = ?`

	obj := models.Car{}

	err := cr.DB.QueryRow(query, id).Scan(
		&obj.ID,
		&obj.UserID,
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
