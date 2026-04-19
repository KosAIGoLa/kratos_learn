package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AdminID  uint32 `json:"admin_id"`
	Username string `json:"username"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

// JWTManager JWT 管理器
type JWTManager struct {
	secret             []byte
	accessTokenExpire  time.Duration
	refreshTokenExpire time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secret string, accessTokenExpire, refreshTokenExpire time.Duration) *JWTManager {
	return &JWTManager{
		secret:             []byte(secret),
		accessTokenExpire:  accessTokenExpire,
		refreshTokenExpire: refreshTokenExpire,
	}
}

// GenerateToken 生成访问 token
func (j *JWTManager) GenerateToken(adminID uint32, username string) (string, error) {
	return j.generateToken(AccessTokenType, adminID, username, j.accessTokenExpire)
}

// GenerateRefreshToken 生成刷新 token
func (j *JWTManager) GenerateRefreshToken(adminID uint32, username string) (string, error) {
	return j.generateToken(RefreshTokenType, adminID, username, j.refreshTokenExpire)
}

func (j *JWTManager) generateToken(tokenType string, adminID uint32, username string, expireAfter time.Duration) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(expireAfter)

	claims := Claims{
		AdminID:  adminID,
		Username: username,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			NotBefore: jwt.NewNumericDate(nowTime),
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString(j.secret)
}

// ParseAccessToken 解析访问 token
func (j *JWTManager) ParseAccessToken(tokenString string) (*Claims, error) {
	return j.parseTokenByType(tokenString, AccessTokenType)
}

func (j *JWTManager) parseTokenByType(tokenString, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.Type != expectedType {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
