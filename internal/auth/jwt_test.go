package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "valid bearer token",
			header:   "Bearer abc123",
			expected: "abc123",
		},
		{
			name:     "no authorization header",
			header:   "",
			expected: "",
		},
		{
			name:     "invalid format",
			header:   "Basic abc123",
			expected: "",
		},
		{
			name:     "bearer lowercase",
			header:   "bearer xyz789",
			expected: "xyz789",
		},
		{
			name:     "token with spaces",
			header:   "Bearer   token123",
			expected: "  token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			got := GetBearerToken(req)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestClaimsStructure(t *testing.T) {
	claims := Claims{
		Roles:   []string{"user", "admin"},
		Country: "US",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	if len(claims.Roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(claims.Roles))
	}
	if claims.Country != "US" {
		t.Errorf("expected country US, got %s", claims.Country)
	}
	if claims.Subject != "user123" {
		t.Errorf("expected subject user123, got %s", claims.Subject)
	}
}

func TestParseTokenWithoutSecret(t *testing.T) {
	// Save and restore original JWT_SECRET
	originalSecret := os.Getenv("JWT_SECRET")
	defer func() {
		if originalSecret != "" {
			os.Setenv("JWT_SECRET", originalSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()

	// Unset JWT_SECRET
	os.Unsetenv("JWT_SECRET")

	_, err := ParseToken("sometoken")
	if err == nil {
		t.Error("expected error when JWT_SECRET not set")
	}
	if err.Error() != "JWT_SECRET not set" {
		t.Errorf("expected 'JWT_SECRET not set', got %v", err)
	}
}

func TestParseTokenWithValidToken(t *testing.T) {
	// Set JWT_SECRET for test
	secret := "test-secret-key-12345"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	// Create a valid token
	claims := &Claims{
		Roles:   []string{"user"},
		Country: "US",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Parse the token
	parsedClaims, err := ParseToken(tokenString)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if parsedClaims.Subject != "user123" {
		t.Errorf("expected subject user123, got %s", parsedClaims.Subject)
	}
	if parsedClaims.Country != "US" {
		t.Errorf("expected country US, got %s", parsedClaims.Country)
	}
	if len(parsedClaims.Roles) != 1 || parsedClaims.Roles[0] != "user" {
		t.Errorf("expected roles [user], got %v", parsedClaims.Roles)
	}
}

func TestParseTokenWithInvalidToken(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	// Test with invalid token string
	_, err := ParseToken("invalid.token.string")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParseTokenWithExpiredToken(t *testing.T) {
	secret := "test-secret-key-12345"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	// Create an expired token
	claims := &Claims{
		Roles:   []string{"user"},
		Country: "US",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Parse the expired token
	_, err = ParseToken(tokenString)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestParseTokenWithWrongSecret(t *testing.T) {
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"

	// Create token with correct secret
	claims := &Claims{
		Roles:   []string{"user"},
		Country: "US",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(correctSecret))
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Try to parse with wrong secret
	os.Setenv("JWT_SECRET", wrongSecret)
	defer os.Unsetenv("JWT_SECRET")

	_, err = ParseToken(tokenString)
	if err == nil {
		t.Error("expected error when using wrong secret")
	}
}
