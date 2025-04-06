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
	pb "xelbot.com/auto-notes/server/proto"
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
		fr.app.ServerError(err)

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
				Id:        int32(dbFuel.Station.ID),
				Name:      dbFuel.Station.Name,
				CreatedAt: timestamppb.New(dbFuel.Station.CreatedAt),
			},
			Date:      timestamppb.New(dbFuel.Date),
			CreatedAt: timestamppb.New(dbFuel.CreatedAt),
		}

		if dbFuel.Distance.Valid {
			fuel.Distance = dbFuel.Distance.Int32
		}

		fuels = append(fuels, fuel)
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
		fr.app.ServerError(err)

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

	currencyRepo := repository.CurrencyRepository{DB: fr.app.DB}
	_, err = currencyRepo.GetCurrencyByCode(currencyCode)
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

	return &pb.Fuel{}, nil
}
