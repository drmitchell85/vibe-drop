package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Service string `json:"service"`
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "File upload endpoint not yet implemented",
		Service: "file-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "File download for ID " + fileID + " not yet implemented",
		Service: "file-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func GetFileMetadataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "File metadata for ID " + fileID + " not yet implemented",
		Service: "file-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "File listing not yet implemented",
		Service: "file-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	response := ErrorResponse{
		Error:   "not_implemented",
		Message: "File deletion for ID " + fileID + " not yet implemented",
		Service: "file-service",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}