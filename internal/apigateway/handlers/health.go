package handlers

import (
	"net/http"
	"time"
	"vibe-drop/internal/common"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "api-gateway",
	}

	common.WriteOKResponse(w, response)
}