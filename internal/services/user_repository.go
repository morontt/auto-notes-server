package services

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/proto"
)

type UserRepositoryService struct {
	app application.Container
}

func NewUserRepositoryService(app application.Container) *UserRepositoryService {
	return &UserRepositoryService{app: app}
}

func (ur *UserRepositoryService) GetCars(ctx context.Context, _ *emptypb.Empty) (*proto.CarCollection, error) {
	return nil, nil
}
