package repository

import (
	"database/sql"
	"errors"

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
			us.created_at,
			us.updated_at
		FROM user_settings AS us
		LEFT JOIN cars AS c ON us.default_car_id = c.id
		WHERE (us.user_id = ?)`

	userSetting := models.UserSetting{}

	err := usr.DB.QueryRow(query, userID).Scan(
		&userSetting.ID,
		&userSetting.CarID,
		&userSetting.CarBrand,
		&userSetting.CarModel,
		&userSetting.CurrencyID,
		&userSetting.CreatedAt,
		&userSetting.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.RecordNotFound
		} else {
			return nil, err
		}
	}

	return &userSetting, nil
}
