package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint32 `json:"user_id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret             []byte
	accessTokenExpire  time.Duration
	refreshTokenExpire time.Duration
}

// NewJWTManager 創建 JWT 管理器
func NewJWTManager(secret string, accessTokenExpire, refreshTokenExpire time.Duration) *JWTManager {
	return &JWTManager{
		secret:             []byte(secret),
		accessTokenExpire:  accessTokenExpire,
		refreshTokenExpire: refreshTokenExpire,
	}
}

// GenerateToken 生成訪問 token
func (j *JWTManager) GenerateToken(userID uint32, username, phone string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(j.accessTokenExpire)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Phone:    phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			NotBefore: jwt.NewNumericDate(nowTime),
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(j.secret)
	return token, err
}

// GenerateRefreshToken 生成刷新 token
func (j *JWTManager) GenerateRefreshToken(userID uint32, username, phone string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(j.refreshTokenExpire)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Phone:    phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			NotBefore: jwt.NewNumericDate(nowTime),
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(j.secret)
	return token, err
}

// ParseToken 解析 token
func (j *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
