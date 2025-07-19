package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	secretKey := "test-secret-key"
	expirationHours := 24

	tm := NewTokenManager(secretKey, expirationHours)

	assert.NotNil(t, tm)
	assert.Equal(t, []byte(secretKey), tm.secretKey)
	assert.Equal(t, time.Duration(expirationHours)*time.Hour, tm.expiration)
}

func TestTokenManager_GenerateToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24)
	userID := uint(123)
	email := "test@example.com"

	token, err := tm.GenerateToken(userID, email)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".")
}

func TestTokenManager_ValidateToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24)
	userID := uint(123)
	email := "test@example.com"

	// Generate a valid token
	token, err := tm.GenerateToken(userID, email)
	require.NoError(t, err)

	// Validate the token
	claims, err := tm.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestTokenManager_ValidateToken_InvalidToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "invalid.token.here",
		},
		{
			name:  "token with wrong signature",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSJ9.wrong_signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tm.ValidateToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Equal(t, ErrInvalidToken, err)
		})
	}
}

func TestTokenManager_ValidateToken_ExpiredToken(t *testing.T) {
	// Create a token manager with very short expiration
	tm := NewTokenManager("test-secret", 0) // 0 hours = immediate expiration
	tm.expiration = -time.Hour              // Set to past time to ensure expiration

	userID := uint(123)
	email := "test@example.com"

	token, err := tm.GenerateToken(userID, email)
	require.NoError(t, err)

	// Wait a moment to ensure token is expired
	time.Sleep(10 * time.Millisecond)

	claims, err := tm.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrExpiredToken, err)
}

func TestTokenManager_ValidateToken_DifferentSecret(t *testing.T) {
	tm1 := NewTokenManager("secret1", 24)
	tm2 := NewTokenManager("secret2", 24)

	userID := uint(123)
	email := "test@example.com"

	// Generate token with first manager
	token, err := tm1.GenerateToken(userID, email)
	require.NoError(t, err)

	// Try to validate with second manager (different secret)
	claims, err := tm2.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestTokenManager_ParseToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24)
	userID := uint(123)
	email := "test@example.com"

	// Generate a token
	token, err := tm.GenerateToken(userID, email)
	require.NoError(t, err)

	// Parse the token without validation
	claims, err := tm.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestTokenManager_ParseToken_InvalidToken(t *testing.T) {
	tm := NewTokenManager("test-secret", 24)

	claims, err := tm.ParseToken("invalid.token")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestTokenManager_GetTokenExpiration(t *testing.T) {
	expirationHours := 48
	tm := NewTokenManager("test-secret", expirationHours)

	expiration := tm.GetTokenExpiration()
	assert.Equal(t, time.Duration(expirationHours)*time.Hour, expiration)
}

func TestClaims_Structure(t *testing.T) {
	tm := NewTokenManager("test-secret", 24)
	userID := uint(456)
	email := "user@test.com"

	token, err := tm.GenerateToken(userID, email)
	require.NoError(t, err)

	claims, err := tm.ValidateToken(token)
	require.NoError(t, err)

	// Verify all claim fields are properly set
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
	assert.True(t, claims.ExpiresAt.After(claims.IssuedAt.Time))
}