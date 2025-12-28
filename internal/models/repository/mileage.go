package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/filters"
	"xelbot.com/auto-notes/server/internal/utils/database"
)

type MileageRepository struct {
	DB *database.DB
}

func (mr *MileageRepository) GetMileagesByUser(userID uint, filter *filters.MileageFilter) ([]*models.Mileage, int, error) {
	cntDs := milageListQueryExpression(userID, filter)
	cntDs = cntDs.ClearSelect().Select(goqu.COUNT("m.id"))

	var count int
	cntQuery, cntParams, _ := cntDs.Prepared(true).ToSQL()
	err := mr.DB.QueryRow(cntQuery, cntParams...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	ds := milageListQueryExpression(userID, filter)
	ds = ds.Order(goqu.I("m.date").Desc(), goqu.I("m.id").Desc())

	if filter.GetLimit() > 0 {
		ds = ds.Limit(uint(filter.GetLimit()))
		if filter.GetPage() > 1 {
			ds = ds.Offset(uint(filter.GetLimit() * (filter.GetPage() - 1)))
		}
	}

	query, params, _ := ds.Prepared(true).ToSQL()
	rows, err := mr.DB.Query(query, params...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	items := make([]*models.Mileage, 0)

	for rows.Next() {
		obj := models.Mileage{}
		carFields := struct {
			ID    sql.NullInt32
			Brand sql.NullString
			Model sql.NullString
		}{}
		err = rows.Scan(
			&obj.ID,
			&obj.Date,
			&obj.Distance,
			&carFields.ID,
			&carFields.Brand,
			&carFields.Model,
			&obj.CreatedAt)

		if err != nil {
			return nil, 0, err
		}

		if carFields.ID.Valid {
			car := models.Car{
				ID:    uint(carFields.ID.Int32),
				Brand: carFields.Brand.String,
				Model: carFields.Model.String,
			}
			obj.Car = &car
		}

		items = append(items, &obj)
	}

	return items, count, nil
}

func (mr *MileageRepository) FindUniq(distance, carId uint, dt time.Time) (*models.Mileage, error) {
	ds := mileageQueryExpression()

	ds = ds.Where(goqu.Ex{
		"m.distance": distance,
		"m.date":     dt.Format(time.DateOnly),
		"c.id":       carId,
	})

	return mr.findMileageRow(ds)
}

func (mr *MileageRepository) Find(id uint) (*models.Mileage, error) {
	ds := mileageQueryExpression()

	ds = ds.Where(goqu.Ex{"m.id": id})

	return mr.findMileageRow(ds)
}

func (mr *MileageRepository) findMileageRow(ds *goqu.SelectDataset) (*models.Mileage, error) {
	query, params, _ := ds.Prepared(true).ToSQL()

	obj := models.Mileage{}
	carFields := struct {
		ID    sql.NullInt32
		Brand sql.NullString
		Model sql.NullString
	}{}

	err := mr.DB.QueryRow(query, params...).Scan(
		&obj.ID,
		&obj.Date,
		&obj.Distance,
		&carFields.ID,
		&carFields.Brand,
		&carFields.Model,
		&obj.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	if carFields.ID.Valid {
		car := models.Car{
			ID:    uint(carFields.ID.Int32),
			Brand: carFields.Brand.String,
			Model: carFields.Model.String,
		}
		obj.Car = &car
	}

	return &obj, nil
}

func (mr *MileageRepository) Validate(obj *models.Mileage) error {
	var dist sql.NullInt32
	err := mr.DB.QueryRow(
		"SELECT MIN(`distance`) FROM mileages WHERE `car_id` = ? AND `date` > ?",
		obj.Car.ID,
		obj.Date.Format(time.DateOnly),
	).Scan(&dist)
	if err != nil {
		return err
	}

	if dist.Valid && uint(dist.Int32) < obj.Distance {
		return models.InvalidMileage
	}

	err = mr.DB.QueryRow(
		"SELECT MAX(`distance`) FROM mileages WHERE `car_id` = ? AND `date` < ?",
		obj.Car.ID,
		obj.Date.Format(time.DateOnly),
	).Scan(&dist)
	if err != nil {
		return err
	}

	if dist.Valid && uint(dist.Int32) > obj.Distance {
		return models.InvalidMileage
	}

	return nil
}

func (mr *MileageRepository) SaveMileage(obj *models.Mileage) (uint, error) {
	data := goqu.Record{}

	data["date"] = obj.Date.Format(time.DateOnly)
	data["distance"] = obj.Distance

	if obj.Car != nil {
		data["car_id"] = obj.Car.ID
	}

	var ds exp.SQLExpression
	if obj.ID == 0 {
		ds = goqu.Dialect("mysql8").Insert("mileages").Rows(data)
	} else {
		ds = goqu.Dialect("mysql8").Update("mileages").Set(data).Where(goqu.Ex{"id": obj.ID})
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return 0, err
	}

	res, err := mr.DB.Exec(query)
	if err != nil {
		return 0, err
	}

	if obj.ID == 0 {
		lastID, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}

		return uint(lastID), nil
	}

	return obj.ID, nil
}

func milageListQueryExpression(userID uint, _ *filters.MileageFilter) *goqu.SelectDataset {
	ds := mileageQueryExpression()

	ds = ds.Where(goqu.Ex{
		"c.user_id": userID,
	})

	return ds
}

func mileageQueryExpression() *goqu.SelectDataset {
	return goqu.Dialect("mysql8").From(goqu.T("mileages").As("m")).Select(
		"m.id",
		goqu.I("m.date").As("m_date"),
		"m.distance",
		goqu.I("c.id").As("car_id"),
		goqu.I("c.brand_name").As("car_brand"),
		goqu.I("c.model_name").As("car_model"),
		"m.created_at",
	).InnerJoin(
		goqu.T("cars").As("c"),
		goqu.On(goqu.Ex{
			"c.id": goqu.I("m.car_id"),
		}),
	)
}
