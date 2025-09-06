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

type FuelRepositoryService struct {
	app application.Container
}

func NewFuelRepositoryService(app application.Container) *FuelRepositoryService {
	return &FuelRepositoryService{app: app}
}

func (fr *FuelRepositoryService) GetFuels(ctx context.Context, limit *pb.Limit) (*pb.FuelCollection, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.FuelRepository{DB: fr.app.DB}
	dbFuels, err := repo.GetFuelsByUser(user.ID, uint(limit.Limit))
	if err != nil {
		fr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	fuels := make([]*pb.Fuel, 0, len(dbFuels))
	for _, dbFuel := range dbFuels {
		fuels = append(fuels, dbFuel.ToRpcMessage())
	}

	fr.app.Info("FuelRepositoryService: populate fuels", ctx, "cnt", len(dbFuels))

	return &pb.FuelCollection{Fuels: fuels}, nil
}

func (fr *FuelRepositoryService) GetFillingStations(ctx context.Context, _ *emptypb.Empty) (*pb.FillingStationCollection, error) {
	_, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.FuelRepository{DB: fr.app.DB}
	dbStations, err := repo.GetFillingStations()
	if err != nil {
		fr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	stations := make([]*pb.FillingStation, 0, len(dbStations))
	for _, dbItem := range dbStations {
		item := &pb.FillingStation{
			Id:        int32(dbItem.ID),
			Name:      dbItem.Name,
			CreatedAt: timestamppb.New(dbItem.CreatedAt),
		}

		stations = append(stations, item)
	}

	fr.app.Info("FuelRepositoryService: populate stations", ctx, "cnt", len(dbStations))

	return &pb.FillingStationCollection{Stations: stations}, nil
}

func (fr *FuelRepositoryService) SaveFuel(ctx context.Context, fuel *pb.Fuel) (*pb.Fuel, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	currencyCode := fuel.Cost.GetCurrency()
	if currencyCode == "" {
		return nil, twirp.InvalidArgument.Error("empty currency code")
	}

	if fuel.Station.GetId() == 0 {
		return nil, twirp.InvalidArgument.Error("empty filling station")
	}

	fuelRepo := repository.FuelRepository{DB: fr.app.DB}
	if fuel.GetId() > 0 {
		ownerId, err := fuelRepo.FuelOwner(uint(fuel.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.NotFound.Error("fuel not found")
			} else {
				return nil, twirp.InternalError("internal error")
			}
		}
		if ownerId != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid fuel owner")
		}
	}

	currencyRepo := repository.CurrencyRepository{DB: fr.app.DB}
	currency, err := currencyRepo.GetCurrencyByCode(currencyCode)
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			return nil, twirp.InvalidArgument.Error("invalid currency")
		} else {
			return nil, twirp.InternalError("internal error")
		}
	}

	carRepo := repository.CarRepository{DB: fr.app.DB}
	car, err := carRepo.Find(uint(fuel.Car.GetId()))
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			return nil, twirp.InvalidArgument.Error("invalid car")
		} else {
			return nil, twirp.InternalError("internal error")
		}
	}

	if car.UserID != user.ID {
		return nil, twirp.InvalidArgument.Error("invalid car owner")
	}

	fuelModel := models.Fuel{
		ID:  uint(fuel.GetId()),
		Car: car,
		Cost: models.Cost{
			Value:      fuel.Cost.GetValue(),
			CurrencyID: currency.ID,
		},
		Value: fuel.GetValue(),
		Date:  fuel.Date.AsTime(),
		Station: models.FillingStation{
			ID: uint(fuel.Station.GetId()),
		},
	}

	fuelID, err := fuelRepo.SaveFuel(&fuelModel)
	if err != nil {
		fr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	dbFuel, err := fuelRepo.FindFuel(fuelID)
	if err != nil {
		fr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return dbFuel.ToRpcMessage(), nil
}
