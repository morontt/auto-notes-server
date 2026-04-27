package server

import (
	"context"
	"errors"

	"github.com/twitchtv/twirp"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/filters"
	"xelbot.com/auto-notes/server/internal/models/repository"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

type CarRepositoryService struct {
	app application.Container
}

func NewCarRepositoryService(app application.Container) *CarRepositoryService {
	return &CarRepositoryService{app: app}
}

func (cr *CarRepositoryService) GetServices(ctx context.Context, pbFilter *pb.ServiceFilter) (*pb.ServiceCollection, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	filter := filters.NewServiceFilter(pbFilter)

	repo := repository.ServiceRepository{DB: cr.app.DB}
	dbItems, cntItems, err := repo.GetServicesByUser(user.ID, filter)
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	if pageOutOfRange(filter, cntItems) {
		return nil, twirp.NotFoundError("services not found")
	}

	items := make([]*pb.Service, 0, len(dbItems))
	for _, dbItem := range dbItems {
		items = append(items, dbItem.ToRpcMessage())
	}

	cr.app.Info("CarRepositoryService: populate services", ctx, "cnt", len(dbItems))

	return &pb.ServiceCollection{
		Services: items,
		Meta: &pb.PaginationMeta{
			Current: int32(filter.GetPage()),
			Last:    int32(filters.GetLastPage(filter, cntItems)),
		},
	}, nil
}

func (cr *CarRepositoryService) FindService(ctx context.Context, idReq *pb.IdRequest) (*pb.Service, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	repo := repository.ServiceRepository{DB: cr.app.DB}
	if idReq.GetId() > 0 {
		ownerId, err := repo.ServiceOwner(uint(idReq.GetId()))
		if err != nil {
			return nil, toTwirpError(cr.app, err, ctx)
		}
		if ownerId != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid service owner")
		}
	} else {
		return nil, twirp.InvalidArgument.Error("invalid id")
	}

	dbItem, err := repo.Find(uint(idReq.GetId()))
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	return dbItem.ToRpcMessage(), nil
}

func (cr *CarRepositoryService) SaveService(ctx context.Context, service *pb.Service) (*pb.Service, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	if service.GetDate() == nil {
		return nil, twirp.InvalidArgument.Error("date is required")
	}

	var cost *models.Cost
	if service.Cost.GetValue() > 0 {
		currencyCode := service.Cost.GetCurrency()
		if currencyCode == "" {
			return nil, twirp.InvalidArgument.Error("empty currency code")
		}

		currencyRepo := repository.CurrencyRepository{DB: cr.app.DB}
		currency, err := currencyRepo.GetCurrencyByCode(currencyCode)
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.InvalidArgument.Error("invalid currency")
			}

			return nil, toTwirpError(cr.app, err, ctx)
		}

		cost = &models.Cost{
			Value:      service.Cost.GetValue(),
			CurrencyID: currency.ID,
		}
	}

	var car *models.Car
	if service.Car.GetId() > 0 {
		carRepo := repository.CarRepository{DB: cr.app.DB}
		car, err = carRepo.Find(uint(service.Car.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.InvalidArgument.Error("invalid car")
			}

			return nil, toTwirpError(cr.app, err, ctx)
		}

		if car.UserID != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid car owner")
		}
	} else {
		return nil, twirp.InvalidArgument.Error("car is required")
	}

	var mileage *models.Mileage
	if service.Distance > 0 && car != nil && service.GetDate() != nil {
		mileageRepo := repository.MileageRepository{DB: cr.app.DB}
		mileage, err = mileageRepo.FindOrCreate(ctx, uint(service.Distance), car.ID, service.GetDate().AsTime())
		if err != nil {
			return nil, toTwirpError(cr.app, err, ctx)
		}
	}

	serviceModel := models.Service{
		ID:          uint(service.GetId()),
		Car:         car,
		Cost:        cost,
		Description: service.GetDescription(),
		Date:        service.Date.AsTime(),
		Mileage:     mileage,
	}

	repo := repository.ServiceRepository{DB: cr.app.DB}
	serviceID, err := repo.SaveService(ctx, &serviceModel, user.ID)
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	dbItem, err := repo.Find(serviceID)
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	return dbItem.ToRpcMessage(), nil
}

func (cr *CarRepositoryService) GetMileages(ctx context.Context, pbFilter *pb.MileageFilter) (*pb.MileageCollection, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	filter := filters.NewMileageFilter(pbFilter)

	repo := repository.MileageRepository{DB: cr.app.DB}
	dbTypes, cntItems, err := repo.GetMileagesByUser(user.ID, filter)
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	if pageOutOfRange(filter, cntItems) {
		return nil, twirp.NotFoundError("mileages not found")
	}

	items := make([]*pb.Mileage, 0, len(dbTypes))
	for _, dbItem := range dbTypes {
		items = append(items, dbItem.ToRpcMessage())
	}

	cr.app.Info("CarRepositoryService: populate mileages", ctx, "cnt", len(dbTypes))

	return &pb.MileageCollection{
		Mileages: items,
		Meta: &pb.PaginationMeta{
			Current: int32(filter.GetPage()),
			Last:    int32(filters.GetLastPage(filter, cntItems)),
		},
	}, nil
}

func (cr *CarRepositoryService) SaveMileage(ctx context.Context, mileage *pb.Mileage) (*pb.Mileage, error) {
	user, err := userClaimsFromContext(ctx)
	if err != nil {
		return nil, twirp.Unauthenticated.Error(err.Error())
	}

	var car *models.Car
	if mileage.Car.GetId() > 0 {
		carRepo := repository.CarRepository{DB: cr.app.DB}
		car, err = carRepo.Find(uint(mileage.Car.GetId()))
		if err != nil {
			if errors.Is(err, models.RecordNotFound) {
				return nil, twirp.InvalidArgument.Error("invalid car")
			}

			return nil, toTwirpError(cr.app, err, ctx)
		}

		if car.UserID != user.ID {
			return nil, twirp.InvalidArgument.Error("invalid car owner")
		}
	} else {
		return nil, twirp.InvalidArgument.Error("car is required")
	}

	var dbItem *models.Mileage
	mileageRepo := repository.MileageRepository{DB: cr.app.DB}

	if mileage.GetId() == 0 {
		dbItem, err = mileageRepo.FindUniq(
			uint(mileage.GetDistance()),
			car.ID,
			mileage.GetDate().AsTime(),
		)
		if err != nil && !errors.Is(err, models.RecordNotFound) {
			return nil, toTwirpError(cr.app, err, ctx)
		}
		if dbItem != nil {
			return dbItem.ToRpcMessage(), nil
		}
	}

	mileageModel := models.Mileage{
		ID:       uint(mileage.GetId()),
		Car:      car,
		Distance: uint(mileage.GetDistance()),
		Date:     mileage.GetDate().AsTime(),
	}

	err = mileageRepo.Validate(&mileageModel)
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	mileageID, err := mileageRepo.SaveMileage(ctx, &mileageModel)
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	dbItem, err = mileageRepo.Find(mileageID)
	if err != nil {
		return nil, toTwirpError(cr.app, err, ctx)
	}

	return dbItem.ToRpcMessage(), nil
}
