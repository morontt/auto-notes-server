package services

import (
	"context"
	"errors"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/repository"
	"xelbot.com/auto-notes/server/internal/security"
	pb "xelbot.com/auto-notes/server/proto"
)

type UserRepositoryService struct {
	app application.Container
}

func NewUserRepositoryService(app application.Container) *UserRepositoryService {
	return &UserRepositoryService{app: app}
}

func (ur *UserRepositoryService) GetCars(ctx context.Context, _ *emptypb.Empty) (*pb.CarCollection, error) {
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		ur.app.Error("UserRepositoryService: unauthenticated", ctx)

		return nil, twirp.Unauthenticated.Error("unauthenticated")
	}

	repo := repository.CarRepository{DB: ur.app.DB}
	dbCars, err := repo.GetCarsByUser(user.ID)
	if err != nil {
		ur.app.ServerError(err)

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

func (ur *UserRepositoryService) GetFuels(ctx context.Context, limit *pb.Limit) (*pb.FuelCollection, error) {
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		ur.app.Error("UserRepositoryService: unauthenticated", ctx)

		return nil, twirp.Unauthenticated.Error("unauthenticated")
	}

	repo := repository.FuelRepository{DB: ur.app.DB}
	dbFuels, err := repo.GetFuelsByUser(user.ID, uint(limit.Limit))
	if err != nil {
		ur.app.ServerError(err)

		return nil, twirp.InternalError("internal error")
	}

	fuels := make([]*pb.Fuel, 0, len(dbFuels))
	for _, dbFuel := range dbFuels {
		fuel := &pb.Fuel{
			Id: int32(dbFuel.ID),
			Car: &pb.Car{
				Id:   int32(dbFuel.Car.ID),
				Name: dbFuel.Car.Brand + " " + dbFuel.Car.Model,
			},
			Cost: &pb.Cost{
				Value:    dbFuel.Cost.Value,
				Currency: dbFuel.Cost.CurrencyCode,
			},
			Value: dbFuel.Value,
			Station: &pb.FillingStation{
				Id:   int32(dbFuel.Station.ID),
				Name: dbFuel.Station.Name,
			},
			Date:      timestamppb.New(dbFuel.Date),
			CreatedAt: timestamppb.New(dbFuel.CreatedAt),
		}

		if dbFuel.Distance.Valid {
			fuel.Distance = dbFuel.Distance.Int32
		}

		fuels = append(fuels, fuel)
	}

	ur.app.Info("UserRepositoryService: populate fuels", ctx, "cnt", len(dbFuels))

	return &pb.FuelCollection{Fuels: fuels}, nil
}

func (ur *UserRepositoryService) GetCurrencies(ctx context.Context, _ *emptypb.Empty) (*pb.CurrencyCollection, error) {
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		ur.app.Error("UserRepositoryService: unauthenticated", ctx)

		return nil, twirp.Unauthenticated.Error("unauthenticated")
	}

	repo := repository.CurrencyRepository{DB: ur.app.DB}
	dbCurrencies, err := repo.GetCurrencies(user.ID)
	if err != nil {
		ur.app.ServerError(err)

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
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		ur.app.Error("UserRepositoryService: unauthenticated", ctx)

		return nil, twirp.Unauthenticated.Error("unauthenticated")
	}

	repo := repository.CurrencyRepository{DB: ur.app.DB}
	dbCurrencies, err := repo.GetCurrencies(user.ID)
	if err != nil {
		ur.app.ServerError(err)

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
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		ur.app.Error("UserRepositoryService: unauthenticated", ctx)

		return nil, twirp.Unauthenticated.Error("unauthenticated")
	}

	return ur.userSettingsFromDB(ctx, user.ID)
}

func (ur *UserRepositoryService) SaveUserSettings(ctx context.Context, settingsReq *pb.UserSettings) (*pb.UserSettings, error) {
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		ur.app.Error("UserRepositoryService: unauthenticated", ctx)

		return nil, twirp.Unauthenticated.Error("unauthenticated")
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

	err := repo.SaveUserSettings(&settings, user.ID)
	if err != nil {
		ur.app.ServerError(err)

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

		ur.app.ServerError(err)

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
