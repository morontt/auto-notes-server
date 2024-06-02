package service

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	"github.com/kataras/jwt"
	"github.com/twitchtv/twirp"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/repository"
	"xelbot.com/auto-notes/server/internal/security"
	pb "xelbot.com/auto-notes/server/proto"
)

const tokenExpiresDuration = 30 * 24 * time.Hour

type AuthService struct {
	app application.Container
}

func NewAuthService(app application.Container) *AuthService {
	return &AuthService{app: app}
}

func (auth *AuthService) GetToken(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" {
		auth.app.InfoLog.Debug("Auth.GetToken: Username is empty")

		return nil, twirp.InvalidArgument.Error("username is required")
	}
	if req.Password == "" {
		auth.app.InfoLog.Debug("Auth.GetToken: Password is empty")

		return nil, twirp.InvalidArgument.Error("password is required")
	}

	repo := repository.UserRepository{DB: auth.app.DB}
	user, err := repo.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			auth.app.Info("Auth.GetToken: User not found", "username", req.Username)
		} else {
			auth.app.ServerError(err)
		}

		return nil, twirp.InvalidArgument.Error("invalid username or password")
	}

	auth.app.Debug("Get token by user", "user", user)
	passwordHash := security.EncodePassword(req.Password, user.Salt)
	if subtle.ConstantTimeCompare([]byte(passwordHash), []byte(user.PasswordHash)) == 0 {
		auth.app.Info("Auth.GetToken: invalid password", "username", req.Username)

		return nil, twirp.InvalidArgument.Error("invalid username or password")
	}

	tokenData, err := createToken(user)
	if err != nil {
		auth.app.ServerError(err)

		return nil, twirp.InternalError("internal error")
	}

	return &pb.LoginResponse{Token: string(tokenData)}, nil
}

func (auth *AuthService) RefreshToken(context.Context, *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	return nil, twirp.Internal.Error("Not implemented")
}

func createToken(user *models.User) ([]byte, error) {
	secret, err := base64.StdEncoding.DecodeString(application.GetConfig().Secret)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	standardClaims := jwt.Claims{
		Expiry:   now.Add(tokenExpiresDuration).Unix(),
		IssuedAt: now.Unix(),
	}

	claims := security.UserClaims{
		Username: user.Username,
		ID:       user.ID,
	}

	return jwt.Sign(jwt.HS256, secret, claims, standardClaims)
}
