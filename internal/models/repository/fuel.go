package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"xelbot.com/auto-notes/server/internal/models"
)

type FuelRepository struct {
	DB *sql.DB
}

func (fr *FuelRepository) GetFuelsByUser(userID, limit uint) ([]*models.Fuel, error) {
	ds := fuelQueryExpression()

	ds = ds.Where(goqu.Ex{
		"c.user_id": userID,
	}).Order(goqu.I("f.date").Desc())

	if limit > 0 {
		ds = ds.Limit(limit)
	}

	query, params, _ := ds.Prepared(true).ToSQL()
	rows, err := fr.DB.Query(query, params...)
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
	ds := fuelQueryExpression()

	ds = ds.Where(goqu.Ex{"f.id": id})
	query, params, _ := ds.Prepared(true).ToSQL()

	obj := models.Fuel{}

	err := fr.DB.QueryRow(query, params...).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	return &obj, nil
}

func (fr *FuelRepository) FuelOwner(fuelId uint) (uint, error) {
	query := `
		SELECT
			c.user_id
		FROM fuels AS f
		INNER JOIN cars AS c ON f.car_id = c.id
		WHERE f.id = ?`

	var userId uint
	err := fr.DB.QueryRow(query, fuelId).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.RecordNotFound
		} else {
			return 0, err
		}
	}

	return userId, nil
}

func (fr *FuelRepository) SaveFuel(fuel *models.Fuel) (uint, error) {
	data := goqu.Record{}

	data["date"] = fuel.Date.Format(time.DateOnly)
	data["car_id"] = fuel.Car.ID
	data["station_id"] = fuel.Station.ID
	data["currency_id"] = fuel.Cost.CurrencyID
	data["cost"] = fmt.Sprintf("%.2f", 0.01*float64(fuel.Cost.Value))
	data["value"] = fmt.Sprintf("%.2f", 0.01*float64(fuel.Value))

	var ds exp.SQLExpression
	if fuel.ID == 0 {
		ds = goqu.Dialect("mysql8").Insert("fuels").Rows(data)
	} else {
		ds = goqu.Dialect("mysql8").Update("fuels").Set(data).Where(goqu.Ex{"id": fuel.ID})
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return 0, err
	}

	res, err := fr.DB.Exec(query)
	if err != nil {
		return 0, err
	}

	if fuel.ID == 0 {
		lastID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}

		return uint(lastID), nil
	}

	return fuel.ID, nil
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

func fuelQueryExpression() *goqu.SelectDataset {
	return goqu.Dialect("mysql8").From(goqu.T("fuels").As("f")).Select(
		"f.id",
		goqu.L("CAST(CONCAT(f.date, ' 12:00:00') AS DATETIME)").As("f_date"),
		goqu.L("CAST(f.value * 100 AS SIGNED INT)").As("value"),
		goqu.I("azs.id").As("station_id"),
		goqu.I("azs.name").As("station_name"),
		goqu.I("azs.created_at").As("station_created_at"),
		goqu.L("CAST(f.cost * 100 AS SIGNED INT)").As("cost"),
		goqu.I("cur.code").As("curr_code"),
		goqu.I("c.id").As("car_id"),
		goqu.I("c.brand_name").As("car_brand"),
		goqu.I("c.model_name").As("car_model"),
		"m.distance",
		"f.created_at",
	).InnerJoin(
		goqu.T("filling_stations").As("azs"),
		goqu.On(goqu.Ex{
			"azs.id": goqu.I("f.station_id"),
		}),
	).InnerJoin(
		goqu.T("cars").As("c"),
		goqu.On(goqu.Ex{
			"c.id": goqu.I("f.car_id"),
		}),
	).InnerJoin(
		goqu.T("currencies").As("cur"),
		goqu.On(goqu.Ex{
			"cur.id": goqu.I("f.currency_id"),
		}),
	).LeftJoin(
		goqu.T("mileages").As("m"),
		goqu.On(goqu.Ex{
			"m.id": goqu.I("f.mileage_id"),
		}),
	)
}
