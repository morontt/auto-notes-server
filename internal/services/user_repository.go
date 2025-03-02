package services

import (
	"context"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models/repository"
	"xelbot.com/auto-notes/server/internal/security"
	pb "xelbot.com/auto-notes/server/proto"
)

type UserRepositoryService struct {
	app application.Container
}

func NewUserRepositoryService(app application.Container) *UserRepositoryService {
	return &UserRepositoryService{app: app}
}

func (ur *UserRepositoryService) GetCars(ctx context.Context, _ *emptypb.Empty) (*pb.CarCollection, error) {
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		ur.app.Error("UserRepositoryService: unauthenticated", ctx)

		return nil, twirp.Unauthenticated.Error("unauthenticated")
	}

	repo := repository.CarRepository{DB: ur.app.DB}
	dbCars, err := repo.GetCarsByUser(user.ID)
	if err != nil {
		ur.app.ServerError(err)

		return nil, twirp.InternalError("internal error")
	}

	cars := make([]*pb.Car, 0, len(dbCars))
	for _, dbCar := range dbCars {
		car := &pb.Car{
			Id:        int32(dbCar.ID),
			Name:      dbCar.Brand + " " + dbCar.Model,
			Default:   dbCar.Default,
			CreatedAt: timestamppb.New(dbCar.CreatedAt),
		}

		if dbCar.Vin.Valid {
			car.Vin = dbCar.Vin.String
		}
		if dbCar.Year.Valid {
			car.Year = dbCar.Year.Int32
		}

		cars = append(cars, car)
	}

	ur.app.Info("UserRepositoryService: populate cars", ctx, "cnt", len(dbCars))

	return &pb.CarCollection{Cars: cars}, nil
}
