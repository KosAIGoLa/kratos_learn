package service

import (
	"github.com/google/wire"
	"user/internal/conf"
	"user/internal/pkg/captcha"
	"user/internal/pkg/jwt"
)

// NewJWTManagerProvider 創建 JWT 管理器
func NewJWTManagerProvider(c *conf.Jwt) *jwt.JWTManager {
	return jwt.NewJWTManager(c.Secret, c.AccessTokenExpire.AsDuration(), c.RefreshTokenExpire.AsDuration())
}

// NewCaptchaManagerProvider 創建驗證碼管理器
func NewCaptchaManagerProvider() *captcha.CaptchaManager {
	return captcha.NewCaptchaManager()
}

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewUserService, NewKYCService, NewTeamService, NewJWTManagerProvider, NewCaptchaManagerProvider)
