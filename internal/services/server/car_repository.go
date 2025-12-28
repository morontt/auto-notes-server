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

func (cr *CarRepositoryService) GetServices(context.Context, *pb.ServiceFilter) (*pb.ServiceCollection, error) {
	return nil, twirp.InternalError("not implemented")
}

func (cr *CarRepositoryService) SaveService(context.Context, *pb.Service) (*pb.Service, error) {
	return nil, twirp.InternalError("not implemented")
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
		cr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
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
			} else {
				cr.app.ServerError(ctx, err)

				return nil, twirp.InternalError("internal error")
			}
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
			cr.app.ServerError(ctx, err)

			return nil, twirp.InternalError("internal error")
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
		if errors.Is(err, models.InvalidMileage) {
			return nil, twirp.InvalidArgument.Error("invalid distance")
		}

		cr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	mileageID, err := mileageRepo.SaveMileage(&mileageModel)
	if err != nil {
		cr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	dbItem, err = mileageRepo.Find(mileageID)
	if err != nil {
		cr.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return dbItem.ToRpcMessage(), nil
}
