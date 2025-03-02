package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"xelbot.com/auto-notes/server/internal/application"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := generateRequestId()
		w.Header().Add("X-Request-Id", reqID)

		ctx := context.WithValue(r.Context(), application.CtxKeyRequestID, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestId() string {
	var buf [13]byte
	var b64 string
	for len(b64) < 10 {
		_, _ = rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	return b64[0:10]
}
