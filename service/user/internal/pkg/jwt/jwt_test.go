package jwt

import (
	"testing"
	"time"
)

func TestTokenTypesAreSeparated(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Minute, 2*time.Minute)

	accessToken, err := manager.GenerateToken(1, "alice", "13800000000")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	refreshToken, err := manager.GenerateRefreshToken(1, "alice", "13800000000")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if _, err := manager.ParseAccessToken(accessToken); err != nil {
		t.Fatalf("ParseAccessToken(accessToken) error = %v", err)
	}
	if _, err := manager.ParseRefreshToken(refreshToken); err != nil {
		t.Fatalf("ParseRefreshToken(refreshToken) error = %v", err)
	}
	if _, err := manager.ParseRefreshToken(accessToken); err == nil {
		t.Fatal("ParseRefreshToken(accessToken) expected error, got nil")
	}
	if _, err := manager.ParseAccessToken(refreshToken); err == nil {
		t.Fatal("ParseAccessToken(refreshToken) expected error, got nil")
	}
}
