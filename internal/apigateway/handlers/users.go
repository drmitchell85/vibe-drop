package handlers

import (
	"net/http"
	"vibe-drop/internal/common"

	"github.com/gorilla/mux"
)

func GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	common.WriteErrorResponse(w, http.StatusNotImplemented, common.ErrorCode("NOT_IMPLEMENTED"), 
		"User profile endpoint not yet implemented", "User profile for ID " + userID + " will be available in a future release")
}

func UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	common.WriteErrorResponse(w, http.StatusNotImplemented, common.ErrorCode("NOT_IMPLEMENTED"), 
		"User profile update endpoint not yet implemented", "User profile update for ID " + userID + " will be available in a future release")
}

func GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	common.WriteErrorResponse(w, http.StatusNotImplemented, common.ErrorCode("NOT_IMPLEMENTED"), 
		"Current user endpoint not yet implemented", "This feature will be available in a future release")
}