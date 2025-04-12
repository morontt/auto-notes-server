package services

import (
	"context"
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

func (auth *AuthService) GetToken(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" {
		auth.app.Debug("Auth.GetToken: Username is empty", ctx)

		return nil, twirp.InvalidArgument.Error("username is required")
	}
	if req.Password == "" {
		auth.app.Debug("Auth.GetToken: Password is empty", ctx)

		return nil, twirp.InvalidArgument.Error("password is required")
	}

	repo := repository.UserRepository{DB: auth.app.DB}
	user, err := repo.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			auth.app.Info("Auth.GetToken: User not found", ctx, "username", req.Username)

			return nil, twirp.InvalidArgument.Error("invalid username or password")
		}

		auth.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	auth.app.Debug("Get token by user", ctx, "user_id", user.ID, "user_name", user.Username)

	if !security.PasswordVerify(req.Password, user.PasswordHash) {
		auth.app.Info("Auth.GetToken: invalid password", ctx, "username", req.Username)

		return nil, twirp.InvalidArgument.Error("invalid username or password")
	}

	return auth.createLoginResponse(ctx, user)
}

func (auth *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	if req.Token == "" {
		auth.app.Debug("Auth.RefreshToken: token is empty", ctx)

		return nil, twirp.InvalidArgument.Error("token is required")
	}

	verifiedToken, err := jwt.Verify(jwt.HS256, application.GetSecretKey(), []byte(req.Token))
	if err != nil {
		auth.app.Debug("Auth.RefreshToken: invalid token", ctx, "err", err.Error())

		return nil, twirp.InvalidArgument.Error("invalid token")
	}

	claims := security.UserClaims{}
	err = verifiedToken.Claims(&claims)
	if err != nil {
		auth.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	auth.app.Debug("Auth.RefreshToken: parsed claims", ctx, "claims", claims)

	repo := repository.UserRepository{DB: auth.app.DB}
	user, err := repo.GetUserByUsername(claims.Username)
	if err != nil {
		if errors.Is(err, models.RecordNotFound) {
			auth.app.Info("Auth.RefreshToken: User not found", ctx, "username", claims.Username)

			return nil, twirp.InvalidArgument.Error("invalid token")
		}

		auth.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return auth.createLoginResponse(ctx, user)
}

func (auth *AuthService) createLoginResponse(ctx context.Context, user *models.User) (*pb.LoginResponse, error) {
	tokenData, err := createToken(user)
	if err != nil {
		auth.app.ServerError(ctx, err)

		return nil, twirp.InternalError("internal error")
	}

	return &pb.LoginResponse{Token: string(tokenData)}, nil
}

func createToken(user *models.User) ([]byte, error) {
	now := time.Now()
	standardClaims := jwt.Claims{
		Expiry:   now.Add(tokenExpiresDuration).Unix(),
		IssuedAt: now.Unix(),
	}

	claims := security.UserClaims{
		Username: user.Username,
		ID:       user.ID,
	}

	return jwt.Sign(jwt.HS256, application.GetSecretKey(), claims, standardClaims)
}
