package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-musthave-diploma-tpl/internal/auth"
)

type userIDContextKey struct{}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := authTokenFromRequest(r)
			subject, err := auth.ParseJWT(token, jwtSecret)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID, err := auth.SubjectToUserID(subject)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) int64 {
	v := ctx.Value(userIDContextKey{})
	if v == nil {
		return 0
	}
	id, _ := v.(int64)
	return id
}

func authTokenFromRequest(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}
