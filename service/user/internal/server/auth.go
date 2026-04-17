package server

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"user/internal/pkg/jwt"
)

// JWTAuth JWT 認證中間件
func JWTAuth(jwtManager *jwt.JWTManager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 從 context 中獲取 transport 信息
			if tr, ok := transport.FromServerContext(ctx); ok {
				// 獲取 Authorization header
				token := tr.RequestHeader().Get("Authorization")
				if token == "" {
					return nil, errors.Unauthorized("UNAUTHORIZED", "missing token")
				}

				// 移除 "Bearer " 前綴
				token = strings.TrimPrefix(token, "Bearer ")

				// 解析 token
				claims, err := jwtManager.ParseToken(token)
				if err != nil {
					return nil, errors.Unauthorized("UNAUTHORIZED", "invalid token")
				}

				// 將用戶信息存入 context
				ctx = context.WithValue(ctx, "user_id", claims.UserID)
				ctx = context.WithValue(ctx, "username", claims.Username)
				ctx = context.WithValue(ctx, "phone", claims.Phone)
			}

			return handler(ctx, req)
		}
	}
}

// GetUserIDFromContext 從 context 中獲取用戶 ID
func GetUserIDFromContext(ctx context.Context) uint32 {
	if userID, ok := ctx.Value("user_id").(uint32); ok {
		return userID
	}
	return 0
}

// GetUsernameFromContext 從 context 中獲取用戶名
func GetUsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value("username").(string); ok {
		return username
	}
	return ""
}
