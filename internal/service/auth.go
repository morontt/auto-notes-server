package service

import (
	"context"
	"github.com/twitchtv/twirp"

	"xelbot.com/auto-notes/server/internal/app"
	pb "xelbot.com/auto-notes/server/proto"
)

type AuthService struct {
	app app.Container
}

func NewAuthService(app app.Container) *AuthService {
	return &AuthService{app: app}
}

func (auth *AuthService) GetToken(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	auth.app.InfoLog.Info("Get token by user: " + req.Username)

	return &pb.LoginResponse{Token: "Hello " + req.Username}, nil
}

func (auth *AuthService) RefreshToken(context.Context, *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	return nil, twirp.Internal.Error("Not implemented")
}
