package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrHashingFailed     = errors.New("password hashing failed")
	ErrVerificationFailed = errors.New("password verification failed")
	ErrInvalidPassword   = errors.New("invalid password")
)

const (
	// DefaultCost is the default bcrypt cost factor
	// Cost of 12 provides good security while maintaining reasonable performance
	DefaultCost = 12
	
	// MinPasswordLength is the minimum allowed password length
	MinPasswordLength = 8
	
	// MaxPasswordLength is the maximum allowed password length
	MaxPasswordLength = 128
)

// Hasher handles password hashing operations
type Hasher struct {
	cost int
}

// NewHasher creates a new password hasher with the default cost
func NewHasher() *Hasher {
	return &Hasher{
		cost: DefaultCost,
	}
}

// NewHasherWithCost creates a new password hasher with a custom cost
func NewHasherWithCost(cost int) *Hasher {
	// Ensure cost is within bcrypt's valid range (4-31)
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	} else if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}
	
	return &Hasher{
		cost: cost,
	}
}

// HashPassword hashes a plain text password using bcrypt
func (h *Hasher) HashPassword(password string) (string, error) {
	// Validate password length
	if len(password) < MinPasswordLength {
		return "", ErrInvalidPassword
	}
	if len(password) > MaxPasswordLength {
		return "", ErrInvalidPassword
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", ErrHashingFailed
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a plain text password against a hashed password
func (h *Hasher) VerifyPassword(hashedPassword, password string) error {
	// Validate password length
	if len(password) < MinPasswordLength {
		return ErrInvalidPassword
	}
	if len(password) > MaxPasswordLength {
		return ErrInvalidPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrVerificationFailed
		}
		return ErrVerificationFailed
	}

	return nil
}

// GetCost returns the current cost factor
func (h *Hasher) GetCost() int {
	return h.cost
}

// Convenience functions for quick usage

// Hash hashes a password using the default hasher
func Hash(password string) (string, error) {
	hasher := NewHasher()
	return hasher.HashPassword(password)
}

// Verify verifies a password using the default hasher
func Verify(hashedPassword, password string) error {
	hasher := NewHasher()
	return hasher.VerifyPassword(hashedPassword, password)
}

// ValidatePasswordStrength validates password strength requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrInvalidPassword
	}
	if len(password) > MaxPasswordLength {
		return ErrInvalidPassword
	}
	
	// Additional strength requirements can be added here
	// For now, we only check length as per the basic requirements
	
	return nil
}