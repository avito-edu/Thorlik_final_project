package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Thorlik/marketplace/internal/service"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	userService service.UserService
	logger      *zap.Logger
}

func NewAuthMiddleware(userService service.UserService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
		logger:      logger,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Unauthorized","message":"Missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"Unauthorized","message":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := m.userService.ValidateToken(token)
		if err != nil {
			m.logger.Warn("Invalid token", zap.Error(err))
			http.Error(w, `{"error":"Unauthorized","message":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value("claims").(*service.Claims)

			allowed := false
			for _, role := range roles {
				if string(claims.Role) == role {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, `{"error":"Forbidden","message":"Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole("admin")(next)
}

func (m *AuthMiddleware) RequireModerator(next http.Handler) http.Handler {
	return m.RequireRole("admin", "moderator")(next)
}
