package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles JWT token creation and validation
type JWTService struct {
	secretKey []byte        // Secret key for signing tokens (keep this safe!)
	expiry    time.Duration // How long tokens are valid
}

// Claims represents the data we store inside JWT tokens
type Claims struct {
	UserID   string `json:"user_id"`   // Which user this token belongs to
	Username string `json:"username"`  // Username for convenience
	jwt.RegisteredClaims                // Standard JWT fields (expiry, issued at, etc.)
}

// NewJWTService creates a new JWT service with the given secret and expiry
func NewJWTService(secretKey string, expiry time.Duration) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey), // Convert string to bytes
		expiry:    expiry,
	}
}

// GenerateToken creates a new JWT token for the given user
func (j *JWTService) GenerateToken(userID, username string) (string, error) {
	// Create the claims (the data we want to store in the token)
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),           // When token was created
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)), // When token expires
			Subject:   userID,                            // Who the token is for
		},
	}

	// Create the token with our claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token with our secret key (this creates the signature)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken checks if a token is valid and returns the user claims
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token and verify the signature
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Make sure the token was signed with the method we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil // Return our secret key for validation
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract the claims from the token
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is valid (not expired, etc.)
	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}

// RefreshToken creates a new token with extended expiry (optional feature)
func (j *JWTService) RefreshToken(oldTokenString string) (string, error) {
	// First validate the old token
	claims, err := j.ValidateToken(oldTokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Create a new token with the same user info but new expiry
	return j.GenerateToken(claims.UserID, claims.Username)
}