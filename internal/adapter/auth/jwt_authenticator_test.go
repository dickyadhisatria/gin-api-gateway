package auth_test

import (
	"gin-api-gateway/internal/adapter/auth"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateToken(secretKey string, userID string, role string, expireIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(expireIn).Unix(),
	})
	return token.SignedString([]byte(secretKey))
}

func TestJWTAuthenticator_Verify(t *testing.T) {
	secret := "test-secret-key"
	authenticator := auth.NewJWTAuthenticator(secret)

	t.Run("valid token", func(t *testing.T) {
		token, err := generateToken(secret, "user-123", "admin", 5*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := authenticator.Verify(token)
		if err != nil {
			t.Fatalf("unexpected verification error: %v", err)
		}

		if claims.UserID != "user-123" {
			t.Errorf("expected user_id 'user-123', got '%s'", claims.UserID)
		}

		if claims.Role != "admin" {
			t.Errorf("expected role 'admin', got '%s'", claims.Role)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := generateToken(secret, "user-123", "admin", -5*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = authenticator.Verify(token)
		if err == nil {
			t.Fatal("expected error for expired token, got nil")
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		token, err := generateToken("wrong-secret", "user-123", "admin", 5*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		_, err = authenticator.Verify(token)
		if err == nil {
			t.Fatal("expected error for invalid signature, got nil")
		}
	})
}
