package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
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
		typeFields := struct {
			ID   sql.NullInt32
			Name sql.NullString
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
			&typeFields.ID,
			&typeFields.Name,
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
		if typeFields.ID.Valid {
			orderType := models.OrderType{
				ID:   uint(typeFields.ID.Int32),
				Name: typeFields.Name.String,
			}
			obj.Type = &orderType
		}

		items = append(items, &obj)
	}

	return items, count, nil
}

func (or *OrderRepository) Find(id uint) (*models.Order, error) {
	ds := orderQueryExpression()

	ds = ds.Where(goqu.Ex{"o.id": id})
	query, params, _ := ds.Prepared(true).ToSQL()

	obj := models.Order{}
	carFields := struct {
		ID    sql.NullInt32
		Brand sql.NullString
		Model sql.NullString
	}{}
	typeFields := struct {
		ID   sql.NullInt32
		Name sql.NullString
	}{}

	err := or.DB.QueryRow(query, params...).Scan(
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
		&typeFields.ID,
		&typeFields.Name,
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
	if typeFields.ID.Valid {
		orderType := models.OrderType{
			ID:   uint(typeFields.ID.Int32),
			Name: typeFields.Name.String,
		}
		obj.Type = &orderType
	}

	return &obj, nil
}

func (or *OrderRepository) OrderOwner(orderId uint) (uint, error) {
	query := `
		SELECT
			user_id
		FROM orders
		WHERE id = ?`

	var userId uint
	err := or.DB.QueryRow(query, orderId).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.RecordNotFound
		} else {
			return 0, err
		}
	}

	return userId, nil
}

func (or *OrderRepository) GetOrderTypes() ([]*models.OrderType, error) {
	query := `
		SELECT
			ot.id,
			ot.name
		FROM order_types AS ot
		ORDER BY ot.name`

	rows, err := or.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]*models.OrderType, 0)

	for rows.Next() {
		obj := models.OrderType{}
		err = rows.Scan(
			&obj.ID,
			&obj.Name)

		if err != nil {
			return nil, err
		}

		items = append(items, &obj)
	}

	return items, nil
}

func (or *OrderRepository) FindType(id uint) (*models.OrderType, error) {
	query := `
		SELECT
			ot.id,
			ot.name
		FROM order_types AS ot
		WHERE ot.id = ?`

	obj := models.OrderType{}

	err := or.DB.QueryRow(query, id).Scan(
		&obj.ID,
		&obj.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	return &obj, nil
}

func (or *OrderRepository) SaveOrder(obj *models.Order, userId uint) (uint, error) {
	data := goqu.Record{}

	data["date"] = obj.Date.Format(time.DateOnly)
	data["currency_id"] = obj.Cost.CurrencyID
	data["description"] = obj.Description
	data["cost"] = fmt.Sprintf("%.2f", 0.01*float64(obj.Cost.Value))

	if obj.Type != nil {
		data["type_id"] = obj.Type.ID
	} else {
		data["type_id"] = nil
	}

	if obj.Car != nil {
		data["car_id"] = obj.Car.ID
	} else {
		data["car_id"] = nil
	}

	if obj.Capacity.Valid {
		data["capacity"] = obj.Capacity.String
	} else {
		data["capacity"] = nil
	}

	if obj.UsedAt.Valid {
		data["used_at"] = obj.UsedAt.Time.Format(time.DateOnly)
	} else {
		data["used_at"] = nil
	}

	var ds exp.SQLExpression
	if obj.ID == 0 {
		data["user_id"] = userId
		ds = goqu.Dialect("mysql8").Insert("orders").Rows(data)
	} else {
		ds = goqu.Dialect("mysql8").Update("orders").Set(data).Where(goqu.Ex{"id": obj.ID})
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return 0, err
	}

	res, err := or.DB.Exec(query)
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
	).LeftJoin(
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
