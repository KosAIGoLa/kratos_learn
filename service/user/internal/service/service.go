package service

import (
	"github.com/google/wire"
	"user/internal/conf"
	"user/internal/pkg/jwt"
)

// NewJWTManagerProvider 創建 JWT 管理器
func NewJWTManagerProvider(c *conf.Jwt) *jwt.JWTManager {
	return jwt.NewJWTManager(c.Secret, c.AccessTokenExpire.AsDuration(), c.RefreshTokenExpire.AsDuration())
}

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewUserService, NewKYCService, NewTeamService, NewJWTManagerProvider)
