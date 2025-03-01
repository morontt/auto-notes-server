package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/kataras/jwt"
	"xelbot.com/auto-notes/server/internal/application"
	"xelbot.com/auto-notes/server/internal/security"
)

func WithAuthorization(app application.Container, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			forbidden(w)
			return
		}

		token, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			forbidden(w)
			return
		}

		verifiedToken, err := jwt.Verify(jwt.HS256, application.GetSecretKey(), []byte(token))
		if err != nil {
			forbidden(w)
			return
		}

		var claims security.UserClaims
		if err = verifiedToken.Claims(&claims); err != nil {
			forbidden(w)
			return
		}

		app.Debug("parsed claims from authorization token", "claims", claims)

		ctx := context.WithValue(r.Context(), security.UserContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func forbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusForbidden)

	_, _ = w.Write([]byte("Who are you?"))
}
