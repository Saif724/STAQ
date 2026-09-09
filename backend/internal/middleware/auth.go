package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/shared/response"
	"github.com/Saif724/STAQ/backend/pkg/jwt"
)

const userIDKey contextKey = "user_id"

func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				response.ErrorJSON(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"authorization header is required",
				)
				return
			}

			parts := strings.Fields(authHeader)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.ErrorJSON(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"invalid authorization header",
				)
				return
			}

			userID, err := jwtManager.VerifyAccessToken(parts[1])
			if err != nil {
				response.ErrorJSON(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"invalid or expired access token",
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				userID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
