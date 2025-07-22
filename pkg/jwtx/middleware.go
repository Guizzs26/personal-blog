package jwtx

import (
	"context"
	"net/http"
	"strings"

	"github.com/Guizzs26/personal-blog/pkg/httpx"
)

type ctxKey string

const UserContextKey ctxKey = "authenticatedUser"

type AuthenticatedUser struct {
	UserID string
	Email  string
}

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "Unauthorized - Missing Bearer Token")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		claims, err := ValidateAccessToken(tokenStr)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "Invalid token")
			return
		}

		user := AuthenticatedUser{
			UserID: claims.UserID,
			Email:  claims.Email,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
