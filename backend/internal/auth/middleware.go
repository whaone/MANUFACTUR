package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"manufactpro/backend/pkg/response"
)

type contextKey string

const (
	ClaimsKey      contextKey = "claims"
	WorkspaceIDKey contextKey = "workspace_id"
	UserIDKey      contextKey = "user_id"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.WriteError(w, http.StatusUnauthorized, "missing token")
			return
		}
		claims, err := ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		ctx = context.WithValue(ctx, WorkspaceIDKey, claims.WorkspaceID)
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetWorkspaceID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(WorkspaceIDKey).(uuid.UUID)
	return v
}

func GetUserID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return v
}
