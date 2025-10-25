package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "User profile for ID " + userID + " not yet implemented",
		Service: "user-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "User profile update for ID " + userID + " not yet implemented",
		Service: "user-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "Current user endpoint not yet implemented",
		Service: "user-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}