package server

import (
	"context"
	"errors"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/repository"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type UserRepositoryService struct {
	app application.Container
}

func NewUserRepositoryService(app application.Container) *UserRepositoryService {
	return &UserRepositoryService{app: app}
}

func (ur *UserRepositoryService) GetCars(ctx context.Context, _ *emptypb.Empty) (*pb.CarCollection, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.CarRepository{DB: ur.app.DB}
	dbCars, err := repo.GetCarsByUser(user.ID)
	if err != nil {
		ur.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	cars := make([]*pb.Car, 0, len(dbCars))
	for _, dbCar := range dbCars {
		car := &pb.Car{
			Id:        int32(dbCar.ID),
			Name:      dbCar.Brand + " " + dbCar.Model,
			Default:   dbCar.Default,
			CreatedAt: timestamppb.New(dbCar.CreatedAt),
		}

		if dbCar.Vin.Valid {
			car.Vin = dbCar.Vin.String
		}
		if dbCar.Year.Valid {
			car.Year = dbCar.Year.Int32
		}

		cars = append(cars, car)
	}

	ur.app.Info("UserRepositoryService: populate cars", ctx, "cnt", len(dbCars))

	return &pb.CarCollection{Cars: cars}, nil
}

func (ur *UserRepositoryService) GetCurrencies(ctx context.Context, _ *emptypb.Empty) (*pb.CurrencyCollection, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.CurrencyRepository{DB: ur.app.DB}
	dbCurrencies, err := repo.GetCurrencies(user.ID)
	if err != nil {
		ur.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	currencies := make([]*pb.Currency, 0, len(dbCurrencies))
	for _, dbCurrency := range dbCurrencies {
		curr := &pb.Currency{
			Id:        int32(dbCurrency.ID),
			Name:      dbCurrency.Name,
			Code:      dbCurrency.Code,
			Default:   dbCurrency.Default,
			CreatedAt: timestamppb.New(dbCurrency.CreatedAt),
		}

		currencies = append(currencies, curr)
	}

	ur.app.Info("UserRepositoryService: populate currencies", ctx, "cnt", len(dbCurrencies))

	return &pb.CurrencyCollection{Currencies: currencies}, nil
}

func (ur *UserRepositoryService) GetDefaultCurrency(ctx context.Context, _ *emptypb.Empty) (*pb.DefaultCurrency, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.CurrencyRepository{DB: ur.app.DB}
	dbCurrencies, err := repo.GetCurrencies(user.ID)
	if err != nil {
		ur.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	result := &pb.DefaultCurrency{}
	for _, dbCurrency := range dbCurrencies {
		if dbCurrency.Default == true {
			result.Currency = &pb.Currency{
				Id:        int32(dbCurrency.ID),
				Name:      dbCurrency.Name,
				Code:      dbCurrency.Code,
				Default:   dbCurrency.Default,
				CreatedAt: timestamppb.New(dbCurrency.CreatedAt),
			}
			result.Found = true

			break
		}
	}

	ur.app.Info("UserRepositoryService: default currency", ctx, "found", result.Found)

	return result, nil
}

func (ur *UserRepositoryService) GetUserSettings(ctx context.Context, _ *emptypb.Empty) (*pb.UserSettings, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	return ur.userSettingsFromDB(ctx, user.ID)
}

func (ur *UserRepositoryService) SaveUserSettings(ctx context.Context, settingsReq *pb.UserSettings) (*pb.UserSettings, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	// TODO check settings owner
	// TODO check car owner

	repo := repository.UserSettingRepository{DB: ur.app.DB}

	settings := models.UserSetting{
		ID: uint(settingsReq.Id),
	}
	if settingsReq.DefaultCar != nil {
		settings.CarID.Valid = true
		settings.CarID.Int32 = settingsReq.DefaultCar.Id
	}
	if settingsReq.DefaultCurrency != nil {
		settings.CurrencyID.Valid = true
		settings.CurrencyID.Int32 = settingsReq.DefaultCurrency.Id
	}

	err = repo.SaveUserSettings(&settings, user.ID)
	if err != nil {
		ur.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return ur.userSettingsFromDB(ctx, user.ID)
}

func (ur *UserRepositoryService) userSettingsFromDB(ctx context.Context, userID uint) (*pb.UserSettings, error) {
	repo := repository.UserSettingRepository{DB: ur.app.DB}
	dbUserSettings, err := repo.GetUserSettings(userID)
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			ur.app.Info("UserRepositoryService: empty user settings", ctx)

			return &pb.UserSettings{}, nil
		}

		ur.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	settings := pb.UserSettings{
		Id:        int32(dbUserSettings.ID),
		CreatedAt: timestamppb.New(dbUserSettings.CreatedAt),
	}

	if dbUserSettings.CarID.Valid {
		settings.DefaultCar = &pb.Car{
			Id:   dbUserSettings.CarID.Int32,
			Name: dbUserSettings.CarBrand.String + " " + dbUserSettings.CarModel.String,
		}
	}
	if dbUserSettings.CurrencyID.Valid {
		settings.DefaultCurrency = &pb.Currency{
			Id:   dbUserSettings.CurrencyID.Int32,
			Name: dbUserSettings.CurrencyName.String,
			Code: dbUserSettings.CurrencyCode.String,
		}
		if dbUserSettings.CurrencyCreatedAt.Valid {
			settings.DefaultCurrency.CreatedAt = timestamppb.New(dbUserSettings.CurrencyCreatedAt.Time)
		}
	}
	if dbUserSettings.UpdatedAt.Valid {
		settings.UpdatedAt = timestamppb.New(dbUserSettings.UpdatedAt.Time)
	}

	ur.app.Info("UserRepositoryService: get user settings", ctx, "user_settings_id", settings.Id)

	return &settings, nil
}
