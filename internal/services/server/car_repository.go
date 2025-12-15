package server

import (
	"context"

	"github.com/twitchtv/twirp"
	"xelbot.com/auto-notes/server/internal/application"
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

func (cr *CarRepositoryService) GetMileages(context.Context, *pb.MileageFilter) (*pb.MileageCollection, error) {
	return nil, twirp.InternalError("not implemented")
}

func (cr *CarRepositoryService) SaveMileage(context.Context, *pb.Mileage) (*pb.Mileage, error) {
	return nil, twirp.InternalError("not implemented")
}
