package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"vibe-drop/internal/auth"
	"vibe-drop/internal/common"
	"vibe-drop/internal/fileservice/storage"
)

// RegisterRequest represents the data sent by client for registration
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents what we send back after successful registration
type RegisterResponse struct {
	User  UserInfo `json:"user"`
	Token string   `json:"token"`
}

// UserInfo represents user data we send to client (no password!)
type UserInfo struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// AuthServices bundles all authentication-related services
type AuthServices struct {
	JWTService      *auth.JWTService
	PasswordService *auth.PasswordService
	DynamoClient    *storage.DynamoClient
}

// RegisterHandler handles user registration
func RegisterHandler(authServices *AuthServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Parse and validate the request
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteValidationError(w, "Invalid request body", err.Error())
			return
		}

		// Step 2: Validate input data using comprehensive validation
		validationReq := &common.UserRegistrationRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
		}
		
		if validationErrors := common.ValidateUserRegistration(validationReq); len(validationErrors) > 0 {
			errorCode, message, details := common.FormatValidationErrors(validationErrors)
			common.WriteErrorResponse(w, http.StatusBadRequest, errorCode, message, details)
			return
		}

		// Step 3: Check if user already exists (by email)
		existingUser, err := authServices.DynamoClient.GetUserByEmail(r.Context(), req.Email)
		if err == nil && existingUser != nil {
			// User exists - don't reveal this for security, but log it
			log.Printf("Registration attempt for existing email: %s", req.Email)
			common.WriteConflictError(w, "User already exists", "A user with this email already exists")
			return
		}

		// Step 4: Hash the password securely
		hashedPassword, err := authServices.PasswordService.HashPassword(req.Password)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			common.WriteInternalServerError(w, "Registration failed", "Unable to process registration")
			return
		}

		// Step 5: Create user record
		userID := uuid.New().String()
		user := &storage.User{
			UserID:       userID,
			Username:     strings.TrimSpace(req.Username),
			Email:        strings.ToLower(strings.TrimSpace(req.Email)),
			PasswordHash: hashedPassword,
		}

		// Step 6: Save user to database
		if err := authServices.DynamoClient.CreateUser(r.Context(), user); err != nil {
			log.Printf("Failed to create user: %v", err)
			common.WriteDatabaseError(w, "Registration failed", "Unable to create user account")
			return
		}

		// Step 7: Generate JWT token for immediate login
		token, err := authServices.JWTService.GenerateToken(user.UserID, user.Username)
		if err != nil {
			log.Printf("Failed to generate token for new user %s: %v", user.UserID, err)
			common.WriteInternalServerError(w, "Registration failed", "Unable to generate access token")
			return
		}

		// Step 8: Return success response with user info and token
		response := RegisterResponse{
			User: UserInfo{
				UserID:    user.UserID,
				Username:  user.Username,
				Email:     user.Email,
				CreatedAt: user.CreatedAt,
			},
			Token: token,
		}

		common.WriteCreatedResponse(w, response)
		log.Printf("Successfully registered new user: %s (%s)", user.Username, user.Email)
	}
}

// LoginRequest represents the data sent by client for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents what we send back after successful login
type LoginResponse struct {
	User  UserInfo `json:"user"`
	Token string   `json:"token"`
}

// LoginHandler handles user login
func LoginHandler(authServices *AuthServices) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Parse the login request
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.WriteValidationError(w, "Invalid request body", err.Error())
			return
		}

		// Step 2: Validate input using comprehensive validation
		if emailErrors := common.ValidateEmail(req.Email); len(emailErrors) > 0 {
			firstError := emailErrors[0]
			common.WriteErrorResponse(w, http.StatusBadRequest, firstError.Code, firstError.Message, 
				fmt.Sprintf("Field: %s", firstError.Field))
			return
		}
		
		if req.Password == "" {
			common.WriteErrorResponse(w, http.StatusBadRequest, common.ErrorCodePasswordRequired, 
				"Password is required", "Field: password")
			return
		}

		// Step 3: Find user by email
		user, err := authServices.DynamoClient.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			// Don't reveal whether user exists or not - security best practice
			log.Printf("Login attempt for non-existent email: %s", req.Email)
			common.WriteUnauthorizedError(w, "Invalid credentials", "Email or password is incorrect")
			return
		}

		// Step 4: Verify password
		err = authServices.PasswordService.VerifyPassword(user.PasswordHash, req.Password)
		if err != nil {
			// Wrong password
			log.Printf("Failed login attempt for user %s: invalid password", user.Email)
			common.WriteUnauthorizedError(w, "Invalid credentials", "Email or password is incorrect")
			return
		}

		// Step 5: Generate JWT token
		token, err := authServices.JWTService.GenerateToken(user.UserID, user.Username)
		if err != nil {
			log.Printf("Failed to generate token for user %s: %v", user.UserID, err)
			common.WriteInternalServerError(w, "Login failed", "Unable to generate access token")
			return
		}

		// Step 6: Return success response
		response := LoginResponse{
			User: UserInfo{
				UserID:    user.UserID,
				Username:  user.Username,
				Email:     user.Email,
				CreatedAt: user.CreatedAt,
			},
			Token: token,
		}

		common.WriteOKResponse(w, response)
		log.Printf("Successful login for user: %s (%s)", user.Username, user.Email)
	}
}


