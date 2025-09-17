package repository

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/filters"
	"xelbot.com/auto-notes/server/internal/utils/database"
)

type OrderRepository struct {
	DB *database.DB
}

func (or *OrderRepository) GetOrdersByUser(userID uint, filter *filters.OrderFilter) ([]*models.Order, int, error) {
	cntDs := orderListQueryExpression(userID, filter)
	cntDs = cntDs.ClearSelect().Select(goqu.COUNT("o.id"))

	var count int
	cntQuery, cntParams, _ := cntDs.Prepared(true).ToSQL()
	err := or.DB.QueryRow(cntQuery, cntParams...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	ds := orderListQueryExpression(userID, filter)
	ds = ds.Order(goqu.I("o.date").Desc(), goqu.I("o.id").Desc())

	if filter.GetLimit() > 0 {
		ds = ds.Limit(uint(filter.GetLimit()))
		if filter.GetPage() > 1 {
			ds = ds.Offset(uint(filter.GetLimit() * (filter.GetPage() - 1)))
		}
	}

	query, params, _ := ds.Prepared(true).ToSQL()
	rows, err := or.DB.Query(query, params...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	items := make([]*models.Order, 0)

	for rows.Next() {
		obj := models.Order{}
		carFields := struct {
			ID    sql.NullInt32
			Brand sql.NullString
			Model sql.NullString
		}{}
		err = rows.Scan(
			&obj.ID,
			&obj.Date,
			&obj.Cost.Value,
			&obj.Cost.CurrencyCode,
			&obj.Description,
			&obj.Capacity,
			&obj.UsedAt,
			&carFields.ID,
			&carFields.Brand,
			&carFields.Model,
			&obj.Distance,
			&obj.Type.ID,
			&obj.Type.Name,
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

func orderListQueryExpression(userID uint, _ *filters.OrderFilter) *goqu.SelectDataset {
	ds := orderQueryExpression()

	ds = ds.Where(goqu.Ex{
		"o.user_id": userID,
	})

	return ds
}

func orderQueryExpression() *goqu.SelectDataset {
	return goqu.Dialect("mysql8").From(goqu.T("orders").As("o")).Select(
		"o.id",
		goqu.I("o.date").As("o_date"),
		goqu.L("CAST(o.cost * 100 AS SIGNED INT)").As("cost"),
		goqu.I("cur.code").As("curr_code"),
		"o.description",
		"o.capacity",
		"o.used_at",
		goqu.I("c.id").As("car_id"),
		goqu.I("c.brand_name").As("car_brand"),
		goqu.I("c.model_name").As("car_model"),
		"m.distance",
		goqu.I("ot.id").As("type_id"),
		goqu.I("ot.name").As("type_name"),
		"o.created_at",
	).LeftJoin(
		goqu.T("cars").As("c"),
		goqu.On(goqu.Ex{
			"c.id": goqu.I("o.car_id"),
		}),
	).InnerJoin(
		goqu.T("currencies").As("cur"),
		goqu.On(goqu.Ex{
			"cur.id": goqu.I("o.currency_id"),
		}),
	).InnerJoin(
		goqu.T("order_types").As("ot"),
		goqu.On(goqu.Ex{
			"ot.id": goqu.I("o.type_id"),
		}),
	).LeftJoin(
		goqu.T("mileages").As("m"),
		goqu.On(goqu.Ex{
			"m.id": goqu.I("o.mileage_id"),
		}),
	)
}
