package server

import (
	"context"
	"strings"

	v1 "user/api/user/v1"
	"user/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

type contextKey string

const (
	userIDContextKey   contextKey = "user_id"
	usernameContextKey contextKey = "username"
	phoneContextKey    contextKey = "phone"
)

var publicOperations = map[string]struct{}{
	v1.OperationUserGetCaptcha:   {},
	v1.OperationUserRegister:     {},
	v1.OperationUserLogin:        {},
	v1.OperationUserRefreshToken: {},
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

			ctx = context.WithValue(ctx, userIDContextKey, claims.UserID)
			ctx = context.WithValue(ctx, usernameContextKey, claims.Username)
			ctx = context.WithValue(ctx, phoneContextKey, claims.Phone)
			return handler(ctx, req)
		}
	}
}

// GetUserIDFromContext 从 context 中获取用户 ID
func GetUserIDFromContext(ctx context.Context) uint32 {
	if userID, ok := ctx.Value(userIDContextKey).(uint32); ok {
		return userID
	}
	return 0
}

// GetUsernameFromContext 从 context 中获取用户名
func GetUsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value(usernameContextKey).(string); ok {
		return username
	}
	return ""
}
