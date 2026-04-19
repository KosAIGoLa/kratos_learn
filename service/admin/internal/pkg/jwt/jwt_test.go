package jwt

import (
	"testing"
	"time"
)

func TestTokenTypesAreSeparated(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Minute, 2*time.Minute)

	accessToken, err := manager.GenerateToken(7, "admin")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	refreshToken, err := manager.GenerateRefreshToken(7, "admin")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if _, err := manager.ParseAccessToken(accessToken); err != nil {
		t.Fatalf("ParseAccessToken(accessToken) error = %v", err)
	}
	if _, err := manager.parseTokenByType(refreshToken, RefreshTokenType); err != nil {
		t.Fatalf("parseTokenByType(refreshToken) error = %v", err)
	}
	if _, err := manager.parseTokenByType(accessToken, RefreshTokenType); err == nil {
		t.Fatal("parseTokenByType(accessToken, refresh) expected error, got nil")
	}
}
