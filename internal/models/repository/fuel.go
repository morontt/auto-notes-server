package repository

import (
	"database/sql"

	"xelbot.com/auto-notes/server/internal/models"
)

type FuelRepository struct {
	DB *sql.DB
}

func (fr *FuelRepository) GetFuelsByUser(userID uint) ([]*models.Fuel, error) {
	query := `
		SELECT
			f.id,
			CAST(CONCAT(f.date, ' 12:00:00') AS DATETIME) AS f_date,
			CAST(f.value * 100 AS SIGNED INT) AS value,
			azs.id AS station_id,
			azs.name AS station_name,
			CAST(f.cost * 100 AS SIGNED INT) AS cost,
			cur.code AS curr_code,
			c.id AS car_id,
			c.brand_name AS car_brand,
			c.model_name AS car_model,
			m.distance,
			f.created_at
		FROM fuels AS f
		INNER JOIN filling_stations AS azs ON f.station_id = azs.id
		INNER JOIN cars AS c ON f.car_id = c.id
		INNER JOIN currencies AS cur ON f.currency_id = cur.id
		LEFT JOIN mileages AS m ON f.mileage_id = m.id
		WHERE c.user_id = ?
		ORDER BY f.date`

	rows, err := fr.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	fuels := make([]*models.Fuel, 0)

	for rows.Next() {
		fuel := models.Fuel{}
		err = rows.Scan(
			&fuel.ID,
			&fuel.Date,
			&fuel.Value,
			&fuel.Station.ID,
			&fuel.Station.Name,
			&fuel.Cost.Value,
			&fuel.Cost.CurrencyCode,
			&fuel.Car.ID,
			&fuel.Car.Brand,
			&fuel.Car.Model,
			&fuel.Distance,
			&fuel.CreatedAt)

		if err != nil {
			return nil, err
		}

		fuels = append(fuels, &fuel)
	}

	return fuels, nil
}
