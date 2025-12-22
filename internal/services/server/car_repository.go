package server

import (
	"context"

	"github.com/twitchtv/twirp"
	"xelbot.com/auto-notes/server/internal/application"
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

func (cr *CarRepositoryService) SaveMileage(context.Context, *pb.Mileage) (*pb.Mileage, error) {
	return nil, twirp.InternalError("not implemented")
}
