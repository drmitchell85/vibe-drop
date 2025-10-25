package handlers

import (
	"encoding/json"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "User login not yet implemented",
		Service: "auth-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "User registration not yet implemented",
		Service: "auth-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "Token refresh not yet implemented",
		Service: "auth-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}