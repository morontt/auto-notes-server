package repository

import (
	"database/sql"

	"xelbot.com/auto-notes/server/internal/models"
)

type FuelRepository struct {
	DB *sql.DB
}

func (fr *FuelRepository) GetFuelsByUser(userID, limit uint) ([]*models.Fuel, error) {
	query := `
		SELECT
			f.id,
			CAST(CONCAT(f.date, ' 12:00:00') AS DATETIME) AS f_date,
			CAST(f.value * 100 AS SIGNED INT) AS value,
			azs.id AS station_id,
			azs.name AS station_name,
			azs.created_at AS station_created_at,
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
		ORDER BY f.date DESC
		LIMIT ?`

	rows, err := fr.DB.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]*models.Fuel, 0)

	for rows.Next() {
		obj := models.Fuel{}
		err = rows.Scan(
			&obj.ID,
			&obj.Date,
			&obj.Value,
			&obj.Station.ID,
			&obj.Station.Name,
			&obj.Station.CreatedAt,
			&obj.Cost.Value,
			&obj.Cost.CurrencyCode,
			&obj.Car.ID,
			&obj.Car.Brand,
			&obj.Car.Model,
			&obj.Distance,
			&obj.CreatedAt)

		if err != nil {
			return nil, err
		}

		items = append(items, &obj)
	}

	return items, nil
}

func (fr *FuelRepository) FindFuel(id uint) (*models.Fuel, error) {
	return nil, nil
}

func (fr *FuelRepository) GetFillingStations() ([]*models.FillingStation, error) {
	query := `
		SELECT
			fs.id,
			fs.name,
			fs.created_at
		FROM filling_stations AS fs
		ORDER BY fs.name`

	rows, err := fr.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]*models.FillingStation, 0)

	for rows.Next() {
		obj := models.FillingStation{}
		err = rows.Scan(
			&obj.ID,
			&obj.Name,
			&obj.CreatedAt)

		if err != nil {
			return nil, err
		}

		items = append(items, &obj)
	}

	return items, nil
}
