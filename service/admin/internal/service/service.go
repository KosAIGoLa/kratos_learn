package service

import (
	"admin/internal/conf"
	"admin/internal/pkg/captcha"
	"admin/internal/pkg/jwt"

	"github.com/google/wire"
)

// NewJWTManagerProvider 创建 JWT 管理器
func NewJWTManagerProvider(c *conf.Jwt) *jwt.JWTManager {
	return jwt.NewJWTManager(c.Secret, c.AccessTokenExpire.AsDuration(), c.RefreshTokenExpire.AsDuration())
}

// NewCaptchaManagerProvider 创建验证码管理器
func NewCaptchaManagerProvider() *captcha.CaptchaManager {
	return captcha.NewCaptchaManager()
}

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewAdminService, NewJWTManagerProvider, NewCaptchaManagerProvider)
