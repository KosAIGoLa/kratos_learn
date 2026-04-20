package server

import (
	"context"
	"strings"

	v1 "admin/api/admin/v1"
	"admin/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

type contextKey string

const (
	adminIDContextKey  contextKey = "admin_id"
	usernameContextKey contextKey = "username"
)

var publicOperations = map[string]struct{}{
	v1.OperationAdminGetCaptcha: {},
	v1.OperationAdminLogin:      {},
}

// JWTAuth JWT 认证中间件
func JWTAuth(jwtManager *jwt.JWTManager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			if _, allowed := publicOperations[tr.Operation()]; allowed {
				return handler(ctx, req)
			}

			token := strings.TrimSpace(tr.RequestHeader().Get("Authorization"))
			if token == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "missing token")
			}

			token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))
			claims, err := jwtManager.ParseAccessToken(token)
			if err != nil {
				return nil, errors.Unauthorized("UNAUTHORIZED", "invalid token")
			}

			ctx = context.WithValue(ctx, adminIDContextKey, claims.AdminID)
			ctx = context.WithValue(ctx, usernameContextKey, claims.Username)
			return handler(ctx, req)
		}
	}
}
