package server

import (
	"context"
	"errors"

	"xelbot.com/auto-notes/server/internal/application"
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
