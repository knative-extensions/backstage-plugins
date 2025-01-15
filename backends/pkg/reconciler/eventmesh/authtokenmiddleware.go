package eventmesh

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type authTokenKey struct{}

const (
	AuthTokenHeader = "Authorization"
	BearerPrefix    = "Bearer "
)

func AuthTokenMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authTokenStr := r.Header.Get(AuthTokenHeader)
			if authTokenStr != "" && strings.HasPrefix(authTokenStr, BearerPrefix) {
				authTokenStr = strings.TrimPrefix(authTokenStr, BearerPrefix)
			}

			// set the token in the context
			if authTokenStr == "" || strings.TrimSpace(authTokenStr) == "" {
				http.Error(w, "missing auth token. set the 'Authorization: Bearer YOUR_KEY' header", http.StatusUnauthorized)
				return
			}

			ctx := WithAuthToken(r.Context(), authTokenStr)

			// Call the handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithAuthToken(ctx context.Context, authToken string) context.Context {
	return context.WithValue(ctx, authTokenKey{}, authToken)
}

func GetAuthToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(authTokenKey{}).(string)
	if token == "" {
		return "", false
	}
	return token, ok
}
