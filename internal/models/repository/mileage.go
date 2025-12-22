package repository

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
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
