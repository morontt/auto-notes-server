package server

import (
	"context"
	"errors"

	"github.com/twitchtv/twirp"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models"
	"xelbot.com/auto-notes/server/internal/models/filters"
	"xelbot.com/auto-notes/server/internal/security"
	pb "xelbot.com/auto-notes/server/rpc/server"
)

var UnAuthenticated = errors.New("services: unauthenticated")

func userClaimsFromContext(ctx context.Context) (*security.UserClaims, error) {
	var (
		user security.UserClaims
		ok   bool
	)

	if user, ok = ctx.Value(application.CtxKeyUser).(security.UserClaims); !ok {
		return nil, UnAuthenticated
	}

	return &user, nil
}

func pageOutOfRange(filter filters.PaginationPart, cntItems int) bool {
	if filter.GetPage() < 1 {
		return true
	}

	if filter.GetPage() > filters.GetLastPage(filter, cntItems) {
		return true
	}

	return false
}

func toTwirpError(app application.Container, err error, ctx context.Context) error {
	if errors.Is(err, models.RecordNotFound) {
		return twirp.NotFound.Error(pb.ErrorCode_E001.String() + ": record not found")
	} else if errors.Is(err, models.InvalidMileage) {
		return twirp.InvalidArgument.Error(pb.ErrorCode_E002.String() + ": invalid distance")
	}

	app.ServerError(ctx, err)

	return twirp.InternalError("internal error")
}
