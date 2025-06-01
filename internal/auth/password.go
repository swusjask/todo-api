package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

// PasswordManager handles password hashing and verification
type PasswordManager struct {
	cost int
}

// NewPasswordManager creates a new password manager
func NewPasswordManager(cost int) *PasswordManager {
	// Default cost if not specified
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordManager{cost: cost}
}

// HashPassword hashes a plain text password
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Generate hash from password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), pm.cost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// CheckPassword compares a plain text password with a hash
func (pm *PasswordManager) CheckPassword(password, hash string) error {
	if password == "" || hash == "" {
		return ErrInvalidPassword
	}

	// Compare password with hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return err
	}

	return nil
}

// ValidatePasswordStrength validates password meets minimum requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// You can add more validation rules here:
	// - Require uppercase letters
	// - Require lowercase letters
	// - Require numbers
	// - Require special characters
	// - Check against common passwords

	hasUpper := false
	hasLower := false
	hasNumber := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	return nil
}
