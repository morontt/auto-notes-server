package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/kataras/jwt"
	"github.com/twitchtv/twirp"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/security"
)

func WithAuthorization(app application.Container, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			app.Warn("Authorization: empty auth header", ctx)

			forbidden(w)
			return
		}

		token, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			app.Warn("Authorization: incorrect auth header", ctx)

			forbidden(w)
			return
		}

		verifiedToken, err := jwt.Verify(jwt.HS256, application.GetSecretKey(), []byte(token))
		if err != nil {
			app.Warn("Authorization: invalid token", ctx, "err", err.Error())

			forbidden(w)
			return
		}

		var claims security.UserClaims
		if err = verifiedToken.Claims(&claims); err != nil {
			app.Warn("Authorization: invalid token claims", ctx, "err", err.Error())

			forbidden(w)
			return
		}

		app.Info("Authorization: parsed claims", ctx, "claims", claims)
		ctx = context.WithValue(ctx, application.CtxKeyUser, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func forbidden(w http.ResponseWriter) {
	twirp.WriteError(w, twirp.PermissionDenied.Error("Who are you?"))
}
