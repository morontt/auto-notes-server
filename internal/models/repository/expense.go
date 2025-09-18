package repository

import (
	"database/sql"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/filters"
	"xelbot.com/auto-notes/server/internal/utils/database"
)

type ExpenseRepository struct {
	DB *database.DB
}

func (er *ExpenseRepository) GetExpensesByUser(userID uint, filter *filters.ExpenseFilter) ([]*models.Expense, int, error) {
	cntDs := expenseListQueryExpression(userID, filter)
	cntDs = cntDs.ClearSelect().Select(goqu.COUNT("e.id"))

	var count int
	cntQuery, cntParams, _ := cntDs.Prepared(true).ToSQL()
	err := er.DB.QueryRow(cntQuery, cntParams...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	ds := expenseListQueryExpression(userID, filter)
	ds = ds.Order(goqu.I("e.date").Desc(), goqu.I("e.id").Desc())

	if filter.GetLimit() > 0 {
		ds = ds.Limit(uint(filter.GetLimit()))
		if filter.GetPage() > 1 {
			ds = ds.Offset(uint(filter.GetLimit() * (filter.GetPage() - 1)))
		}
	}

	query, params, _ := ds.Prepared(true).ToSQL()
	rows, err := er.DB.Query(query, params...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	items := make([]*models.Expense, 0)

	for rows.Next() {
		obj := models.Expense{}
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
			&carFields.ID,
			&carFields.Brand,
			&carFields.Model,
			&obj.Type,
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

func (er *ExpenseRepository) Find(id uint) (*models.Expense, error) {
	ds := expenseQueryExpression()

	ds = ds.Where(goqu.Ex{"e.id": id})
	query, params, _ := ds.Prepared(true).ToSQL()

	obj := models.Expense{}
	carFields := struct {
		ID    sql.NullInt32
		Brand sql.NullString
		Model sql.NullString
	}{}

	err := er.DB.QueryRow(query, params...).Scan(
		&obj.ID,
		&obj.Date,
		&obj.Cost.Value,
		&obj.Cost.CurrencyCode,
		&obj.Description,
		&carFields.ID,
		&carFields.Brand,
		&carFields.Model,
		&obj.Type,
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

func (er *ExpenseRepository) ExpenseOwner(orderId uint) (uint, error) {
	query := `
		SELECT
			user_id
		FROM expenses
		WHERE id = ?`

	var userId uint
	err := er.DB.QueryRow(query, orderId).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.RecordNotFound
		} else {
			return 0, err
		}
	}

	return userId, nil
}

func expenseListQueryExpression(userID uint, _ *filters.ExpenseFilter) *goqu.SelectDataset {
	ds := expenseQueryExpression()

	ds = ds.Where(goqu.Ex{
		"e.user_id": userID,
	})

	return ds
}

func expenseQueryExpression() *goqu.SelectDataset {
	return goqu.Dialect("mysql8").From(goqu.T("expenses").As("e")).Select(
		"e.id",
		goqu.I("e.date").As("e_date"),
		goqu.L("CAST(e.cost * 100 AS SIGNED INT)").As("cost"),
		goqu.I("cur.code").As("curr_code"),
		"e.description",
		goqu.I("c.id").As("car_id"),
		goqu.I("c.brand_name").As("car_brand"),
		goqu.I("c.model_name").As("car_model"),
		"e.type",
		"e.created_at",
	).LeftJoin(
		goqu.T("cars").As("c"),
		goqu.On(goqu.Ex{
			"c.id": goqu.I("e.car_id"),
		}),
	).InnerJoin(
		goqu.T("currencies").As("cur"),
		goqu.On(goqu.Ex{
			"cur.id": goqu.I("e.currency_id"),
		}),
	)
}
