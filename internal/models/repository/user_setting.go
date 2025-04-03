package repository

import (
	"database/sql"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"xelbot.com/auto-notes/server/internal/models"
)

type UserSettingRepository struct {
	DB *sql.DB
}

func (usr *UserSettingRepository) GetUserSettings(userID uint) (*models.UserSetting, error) {
	query := `
		SELECT
			us.id,
			us.default_car_id,
			c.brand_name,
			c.model_name,
			us.default_currency_id,
			cr.name,
			cr.code,
			cr.created_at AS cr_created_at,
			us.created_at,
			us.updated_at
		FROM user_settings AS us
		LEFT JOIN cars AS c ON us.default_car_id = c.id
		LEFT JOIN currencies AS cr ON us.default_currency_id = cr.id
		WHERE (us.user_id = ?)`

	obj := models.UserSetting{}

	err := usr.DB.QueryRow(query, userID).Scan(
		&obj.ID,
		&obj.CarID,
		&obj.CarBrand,
		&obj.CarModel,
		&obj.CurrencyID,
		&obj.CurrencyName,
		&obj.CurrencyCode,
		&obj.CurrencyCreatedAt,
		&obj.CreatedAt,
		&obj.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	return &obj, nil
}

func (usr *UserSettingRepository) SaveUserSettings(settings *models.UserSetting, userId uint) error {
	if settings == nil {
		return errors.New("user settings cannot be null")
	}

	data := goqu.Record{}
	if settings.CarID.Int32 > 0 {
		data["default_car_id"] = settings.CarID.Int32
	} else {
		data["default_car_id"] = nil
	}
	if settings.CurrencyID.Int32 > 0 {
		data["default_currency_id"] = settings.CurrencyID.Int32
	} else {
		data["default_currency_id"] = nil
	}

	var ds exp.SQLExpression
	if settings.ID == 0 {
		data["user_id"] = userId
		ds = goqu.Dialect("mysql8").Insert("user_settings").Rows(data)
	} else {
		ds = goqu.Dialect("mysql8").Update("user_settings").Set(data).Where(goqu.Ex{"user_id": userId})
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return err
	}

	_, err = usr.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
