package handlers

import (
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"vibe-drop/internal/apigateway/services"
	"vibe-drop/internal/common"
)


var fileServiceClient *services.FileServiceClient

func InitializeFileServiceClient(fileServiceURL string) {
	fileServiceClient = services.NewFileServiceClient(fileServiceURL)
}

func getRequestID(r *http.Request) string {
	if id := r.Context().Value("request_id"); id != nil {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

func proxyToFileService(w http.ResponseWriter, r *http.Request, path string) {
	requestID := getRequestID(r)
	
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[%s] Failed to read request body: %v", requestID, err)
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
	
	// Make request to file service
	resp, err := fileServiceClient.ProxyRequest(r.Method, path, body, headers)
	if err != nil {
		log.Printf("[%s] File service request failed: %v", requestID, err)
		common.WriteErrorResponse(w, http.StatusServiceUnavailable, common.ErrorCodeServiceUnavailable, 
			"File service is currently unavailable", err.Error())
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
		log.Printf("[%s] Failed to copy response body: %v", requestID, err)
	}
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	proxyToFileService(w, r, "/files/upload-url")
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]
	proxyToFileService(w, r, "/files/"+fileID+"/download-url")
}

func GetFileMetadataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]
	proxyToFileService(w, r, "/files/"+fileID)
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	proxyToFileService(w, r, "/files")
}

func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]
	proxyToFileService(w, r, "/files/"+fileID)
}