package password

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewHasher(t *testing.T) {
	hasher := NewHasher()
	
	assert.NotNil(t, hasher)
	assert.Equal(t, DefaultCost, hasher.cost)
}

func TestNewHasherWithCost(t *testing.T) {
	tests := []struct {
		name         string
		inputCost    int
		expectedCost int
	}{
		{
			name:         "valid cost",
			inputCost:    10,
			expectedCost: 10,
		},
		{
			name:         "cost too low",
			inputCost:    2,
			expectedCost: bcrypt.MinCost,
		},
		{
			name:         "cost too high",
			inputCost:    50,
			expectedCost: bcrypt.MaxCost,
		},
		{
			name:         "default cost",
			inputCost:    DefaultCost,
			expectedCost: DefaultCost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewHasherWithCost(tt.inputCost)
			assert.Equal(t, tt.expectedCost, hasher.cost)
		})
	}
}

func TestHasher_HashPassword(t *testing.T) {
	hasher := NewHasher()
	password := "testpassword123"

	hashedPassword, err := hasher.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
	assert.True(t, strings.HasPrefix(hashedPassword, "$2a$") || strings.HasPrefix(hashedPassword, "$2b$"))
}

func TestHasher_HashPassword_InvalidLength(t *testing.T) {
	hasher := NewHasher()

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "too short",
			password: "short",
		},
		{
			name:     "empty",
			password: "",
		},
		{
			name:     "too long",
			password: strings.Repeat("a", MaxPasswordLength+1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := hasher.HashPassword(tt.password)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidPassword, err)
			assert.Empty(t, hashedPassword)
		})
	}
}

func TestHasher_VerifyPassword(t *testing.T) {
	hasher := NewHasher()
	password := "testpassword123"

	// Hash the password first
	hashedPassword, err := hasher.HashPassword(password)
	require.NoError(t, err)

	// Verify the correct password
	err = hasher.VerifyPassword(hashedPassword, password)
	assert.NoError(t, err)
}

func TestHasher_VerifyPassword_WrongPassword(t *testing.T) {
	hasher := NewHasher()
	password := "testpassword123"
	wrongPassword := "wrongpassword123"

	// Hash the correct password
	hashedPassword, err := hasher.HashPassword(password)
	require.NoError(t, err)

	// Try to verify with wrong password
	err = hasher.VerifyPassword(hashedPassword, wrongPassword)
	assert.Error(t, err)
	assert.Equal(t, ErrVerificationFailed, err)
}

func TestHasher_VerifyPassword_InvalidLength(t *testing.T) {
	hasher := NewHasher()
	validPassword := "validpassword123"
	
	// Hash a valid password
	hashedPassword, err := hasher.HashPassword(validPassword)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "too short",
			password: "short",
		},
		{
			name:     "empty",
			password: "",
		},
		{
			name:     "too long",
			password: strings.Repeat("a", MaxPasswordLength+1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.VerifyPassword(hashedPassword, tt.password)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidPassword, err)
		})
	}
}

func TestHasher_GetCost(t *testing.T) {
	cost := 10
	hasher := NewHasherWithCost(cost)
	
	assert.Equal(t, cost, hasher.GetCost())
}

func TestHash_ConvenienceFunction(t *testing.T) {
	password := "testpassword123"

	hashedPassword, err := Hash(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
}

func TestVerify_ConvenienceFunction(t *testing.T) {
	password := "testpassword123"

	// Hash using convenience function
	hashedPassword, err := Hash(password)
	require.NoError(t, err)

	// Verify using convenience function
	err = Verify(hashedPassword, password)
	assert.NoError(t, err)

	// Verify with wrong password
	err = Verify(hashedPassword, "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, ErrVerificationFailed, err)
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "validpassword123",
			wantErr:  false,
		},
		{
			name:     "minimum length",
			password: "12345678",
			wantErr:  false,
		},
		{
			name:     "maximum length",
			password: strings.Repeat("a", MaxPasswordLength),
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "empty",
			password: "",
			wantErr:  true,
		},
		{
			name:     "too long",
			password: strings.Repeat("a", MaxPasswordLength+1),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidPassword, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHasher_DifferentCosts(t *testing.T) {
	password := "testpassword123"
	
	// Test different cost factors
	costs := []int{4, 8, 10, 12}
	
	for _, cost := range costs {
		t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
			hasher := NewHasherWithCost(cost)
			
			hashedPassword, err := hasher.HashPassword(password)
			require.NoError(t, err)
			
			err = hasher.VerifyPassword(hashedPassword, password)
			assert.NoError(t, err)
		})
	}
}

func TestHasher_ConsistentHashing(t *testing.T) {
	hasher := NewHasher()
	password := "testpassword123"

	// Hash the same password multiple times
	hash1, err1 := hasher.HashPassword(password)
	hash2, err2 := hasher.HashPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)
	
	// Hashes should be different (due to salt)
	assert.NotEqual(t, hash1, hash2)
	
	// But both should verify correctly
	assert.NoError(t, hasher.VerifyPassword(hash1, password))
	assert.NoError(t, hasher.VerifyPassword(hash2, password))
}