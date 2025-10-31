package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"vibe-drop/internal/common"
)

// UserContextKey is used to store user info in request context
// We use a custom type to avoid collisions with other context keys
type UserContextKey string

const (
	UserIDKey   UserContextKey = "user_id"
	UsernameKey UserContextKey = "username"
)

// AuthMiddleware creates middleware that validates JWT tokens
// This is a "middleware factory" - it returns the actual middleware function
func AuthMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// This is the actual middleware function that gets called for each request
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			
			// Step 1: Extract the token from the Authorization header
			token, err := extractTokenFromHeader(r)
			if err != nil {
				// No valid token found - reject the request
				common.WriteUnauthorizedError(w, "Authentication required", err.Error())
				return // Stop here - don't call next handler
			}
			
			// Step 2: Validate the JWT token
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				// Invalid token - reject the request
				common.WriteUnauthorizedError(w, "Invalid or expired token", err.Error())
				return // Stop here - don't call next handler
			}
			
			// Step 3: Add user info to request context
			// This is how we "pass" the user info to the next handler
			ctx := addUserToContext(r.Context(), claims)
			requestWithUser := r.WithContext(ctx)
			
			// Step 4: Call the next handler with the enhanced request
			// The next handler can now access user info from context
			next.ServeHTTP(w, requestWithUser)
		})
	}
}

// extractTokenFromHeader gets the JWT token from the Authorization header
func extractTokenFromHeader(r *http.Request) (string, error) {
	// Look for: Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	
	// Split "Bearer eyJhbGciOiJIUzI1NiIs..." into ["Bearer", "eyJhbGciOiJIUzI1NiIs..."]
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	
	// Check that it starts with "Bearer"
	if parts[0] != "Bearer" {
		return "", fmt.Errorf("authorization header must start with 'Bearer'")
	}
	
	token := parts[1]
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	
	return token, nil
}

// addUserToContext puts user information into the request context
func addUserToContext(ctx context.Context, claims *Claims) context.Context {
	// Add user ID to context
	ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
	// Add username to context  
	ctx = context.WithValue(ctx, UsernameKey, claims.Username)
	return ctx
}


// Helper functions for handlers to extract user info from context

// GetUserIDFromContext extracts the user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetUsernameFromContext extracts the username from request context
func GetUsernameFromContext(ctx context.Context) (string, error) {
	username, ok := ctx.Value(UsernameKey).(string)
	if !ok {
		return "", fmt.Errorf("username not found in context")
	}
	return username, nil
}

// GetUserFromContext extracts both user ID and username from context
func GetUserFromContext(ctx context.Context) (userID, username string, err error) {
	userID, err = GetUserIDFromContext(ctx)
	if err != nil {
		return "", "", err
	}
	
	username, err = GetUsernameFromContext(ctx)
	if err != nil {
		return "", "", err
	}
	
	return userID, username, nil
}