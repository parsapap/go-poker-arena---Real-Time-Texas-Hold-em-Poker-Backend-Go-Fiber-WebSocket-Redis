package middleware

import (
	"os"
	"testing"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	
	userID := uint(1)
	username := "testuser"
	
	token, err := GenerateToken(userID, username)
	if err != nil {
		t.Errorf("GenerateToken failed: %v", err)
	}
	
	if token == "" {
		t.Error("Token should not be empty")
	}
	
	// Parse and verify token
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	
	if err != nil {
		t.Errorf("Failed to parse token: %v", err)
	}
	
	if !parsedToken.Valid {
		t.Error("Token should be valid")
	}
	
	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}
	
	if claims.Username != username {
		t.Errorf("Expected Username %s, got %s", username, claims.Username)
	}
}

func TestGenerateTokenExpiration(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	
	token, _ := GenerateToken(1, "testuser")
	
	claims := &Claims{}
	jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	
	// Check expiration is ~24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	
	diff := actualExpiry.Sub(expectedExpiry)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("Token expiration not set correctly. Expected ~%v, got %v", expectedExpiry, actualExpiry)
	}
}

func TestGenerateTokenIssuedAt(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	// JWT NumericDate values are stored with whole-second precision, so the
	// issued-at claim is the token's creation time truncated to the second.
	// Compare against second-truncated bounds to avoid a sub-second flake.
	before := time.Now().Truncate(time.Second)
	token, _ := GenerateToken(1, "testuser")
	after := time.Now()

	claims := &Claims{}
	jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})

	issuedAt := claims.IssuedAt.Time
	if issuedAt.Before(before) || issuedAt.After(after.Add(time.Second)) {
		t.Errorf("IssuedAt timestamp not set correctly: issuedAt=%v before=%v after=%v", issuedAt, before, after)
	}
}
