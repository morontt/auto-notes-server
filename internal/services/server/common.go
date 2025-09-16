package server

import (
	"context"
	"errors"

	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/models/filters"
	"xelbot.com/auto-notes/server/internal/security"
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
