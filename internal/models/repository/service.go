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

type ServiceRepository struct {
	DB *database.DB
}

func (sr *ServiceRepository) GetServicesByUser(userID uint, filter *filters.ServiceFilter) ([]*models.Service, int, error) {
	cntDs := serviceListQueryExpression(userID, filter)
	cntDs = cntDs.ClearSelect().Select(goqu.COUNT("s.id"))

	var count int
	cntQuery, cntParams, _ := cntDs.Prepared(true).ToSQL()
	err := sr.DB.QueryRow(cntQuery, cntParams...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	ds := serviceListQueryExpression(userID, filter)
	ds = ds.Order(goqu.I("s.date").Desc(), goqu.I("s.id").Desc())

	if filter.GetLimit() > 0 {
		ds = ds.Limit(uint(filter.GetLimit()))
		if filter.GetPage() > 1 {
			ds = ds.Offset(uint(filter.GetLimit() * (filter.GetPage() - 1)))
		}
	}

	query, params, _ := ds.Prepared(true).ToSQL()
	rows, err := sr.DB.Query(query, params...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	items := make([]*models.Service, 0)

	for rows.Next() {
		obj := models.Service{}
		carFields := struct {
			ID    sql.NullInt32
			Brand sql.NullString
			Model sql.NullString
		}{}
		costFields := struct {
			Value        sql.NullInt32
			CurrencyCode sql.NullString
		}{}
		err = rows.Scan(
			&obj.ID,
			&obj.Date,
			&costFields.Value,
			&costFields.CurrencyCode,
			&obj.Description,
			&carFields.ID,
			&carFields.Brand,
			&carFields.Model,
			&obj.Distance,
			&obj.CreatedAt)

		if err != nil {
			return nil, 0, err
		}

		if carFields.ID.Valid {
			obj.Car = &models.Car{
				ID:    uint(carFields.ID.Int32),
				Brand: carFields.Brand.String,
				Model: carFields.Model.String,
			}
		}
		if costFields.CurrencyCode.Valid {
			obj.Cost = &models.Cost{
				Value:        costFields.Value.Int32,
				CurrencyCode: costFields.CurrencyCode.String,
			}
		}

		items = append(items, &obj)
	}

	return items, count, nil
}

func (sr *ServiceRepository) Find(id uint) (*models.Service, error) {
	ds := orderQueryExpression()

	ds = ds.Where(goqu.Ex{"s.id": id})
	query, params, _ := ds.Prepared(true).ToSQL()

	obj := models.Service{}
	carFields := struct {
		ID    sql.NullInt32
		Brand sql.NullString
		Model sql.NullString
	}{}
	costFields := struct {
		Value        sql.NullInt32
		CurrencyCode sql.NullString
	}{}

	err := sr.DB.QueryRow(query, params...).Scan(
		&obj.ID,
		&obj.Date,
		&costFields.Value,
		&costFields.CurrencyCode,
		&obj.Description,
		&carFields.ID,
		&carFields.Brand,
		&carFields.Model,
		&obj.Distance,
		&obj.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	if carFields.ID.Valid {
		obj.Car = &models.Car{
			ID:    uint(carFields.ID.Int32),
			Brand: carFields.Brand.String,
			Model: carFields.Model.String,
		}
	}
	if costFields.CurrencyCode.Valid {
		obj.Cost = &models.Cost{
			Value:        costFields.Value.Int32,
			CurrencyCode: costFields.CurrencyCode.String,
		}
	}

	return &obj, nil
}

func (sr *ServiceRepository) ServiceOwner(orderId uint) (uint, error) {
	query := `
		SELECT
			c.user_id
		FROM services AS s
		INNER JOIN cars AS c ON c.id = s.car_id
		WHERE s.id = ?`

	var userId uint
	err := sr.DB.QueryRow(query, orderId).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.RecordNotFound
		} else {
			return 0, err
		}
	}

	return userId, nil
}

func (sr *ServiceRepository) SaveService(obj *models.Service, userId uint) (uint, error) {
	data := goqu.Record{}

	data["date"] = obj.Date.Format(time.DateOnly)
	data["description"] = obj.Description

	//data["cost"] = fmt.Sprintf("%.2f", 0.01*float64(obj.Cost.Value))
	//data["currency_id"] = obj.Cost.CurrencyID

	if obj.Car != nil {
		data["car_id"] = obj.Car.ID
	} else {
		data["car_id"] = nil
	}

	if obj.Mileage != nil {
		data["mileage_id"] = obj.Mileage.ID
	} else {
		data["mileage_id"] = nil
	}

	var ds exp.SQLExpression
	if obj.ID == 0 {
		ds = goqu.Dialect("mysql8").Insert("services").Rows(data)
	} else {
		ds = goqu.Dialect("mysql8").Update("services").Set(data).Where(goqu.Ex{"id": obj.ID})
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return 0, err
	}

	res, err := sr.DB.Exec(query)
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

func serviceListQueryExpression(userID uint, _ *filters.ServiceFilter) *goqu.SelectDataset {
	ds := serviceQueryExpression()

	ds = ds.Where(goqu.Ex{
		"c.user_id": userID,
	})

	return ds
}

func serviceQueryExpression() *goqu.SelectDataset {
	return goqu.Dialect("mysql8").From(goqu.T("services").As("s")).Select(
		"s.id",
		goqu.I("s.date").As("s_date"),
		goqu.L("CAST(s.cost * 100 AS SIGNED INT)").As("cost"),
		goqu.I("cur.code").As("curr_code"),
		"s.description",
		goqu.I("c.id").As("car_id"),
		goqu.I("c.brand_name").As("car_brand"),
		goqu.I("c.model_name").As("car_model"),
		"m.distance",
		"s.created_at",
	).InnerJoin(
		goqu.T("cars").As("c"),
		goqu.On(goqu.Ex{
			"c.id": goqu.I("s.car_id"),
		}),
	).LeftJoin(
		goqu.T("currencies").As("cur"),
		goqu.On(goqu.Ex{
			"cur.id": goqu.I("s.currency_id"),
		}),
	).LeftJoin(
		goqu.T("mileages").As("m"),
		goqu.On(goqu.Ex{
			"m.id": goqu.I("s.mileage_id"),
		}),
	)
}
