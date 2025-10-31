package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password hashing and verification
type PasswordService struct {
	cost int // bcrypt cost factor (higher = more secure but slower)
}

// NewPasswordService creates a new password service
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: bcrypt.DefaultCost, // Usually 10, good balance of security vs speed
	}
}

// HashPassword converts a plain text password into a secure hash
func (p *PasswordService) HashPassword(password string) (string, error) {
	// bcrypt automatically handles salt generation and incorporates it into the hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	return string(hashedBytes), nil
}

// VerifyPassword checks if a plain text password matches the hashed password
func (p *PasswordService) VerifyPassword(hashedPassword, plainPassword string) error {
	// bcrypt.CompareHashAndPassword extracts the salt from the hash and compares
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	
	return nil
}

// ValidatePassword checks if a password meets our security requirements
func (p *PasswordService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	
	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters long")
	}
	
	// You could add more rules here:
	// - Require uppercase/lowercase letters
	// - Require numbers
	// - Require special characters
	// - Check against common passwords
	
	return nil
}