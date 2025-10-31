package handlers

import (
	"io"
	"log"
	"net/http"
	"vibe-drop/internal/common"
)

func proxyToFileServiceAuth(w http.ResponseWriter, r *http.Request, path string) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		common.WriteBadRequestError(w, "Failed to read request body", err.Error())
		return
	}
	defer r.Body.Close()
	
	// Copy headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	// Make request to file service (which handles auth)
	resp, err := fileServiceClient.ProxyRequest(r.Method, path, body, headers)
	if err != nil {
		log.Printf("File service auth request failed: %v", err)
		common.WriteErrorResponse(w, http.StatusServiceUnavailable, common.ErrorCodeServiceUnavailable, 
			"Authentication service is currently unavailable", err.Error())
		return
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	
	// Copy status code
	w.WriteHeader(resp.StatusCode)
	
	// Copy response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Failed to copy auth response body: %v", err)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	proxyToFileServiceAuth(w, r, "/auth/login")
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	proxyToFileServiceAuth(w, r, "/auth/register")
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// This endpoint doesn't exist yet in file service, so return not implemented
	common.WriteErrorResponse(w, http.StatusNotImplemented, common.ErrorCode("NOT_IMPLEMENTED"), 
		"Token refresh not yet implemented", "This feature will be available in a future release")
}