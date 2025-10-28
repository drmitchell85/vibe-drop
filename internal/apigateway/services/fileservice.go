package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type FileServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewFileServiceClient(baseURL string) *FileServiceClient {
	return &FileServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *FileServiceClient) ProxyRequest(method, path string, body []byte, headers map[string]string) (*http.Response, error) {
	url := f.baseURL + path
	
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Copy headers from original request
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	// Set default content type if not provided
	if req.Header.Get("Content-Type") == "" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to file service: %w", err)
	}
	
	return resp, nil
}

func (f *FileServiceClient) Health() (*http.Response, error) {
	return f.ProxyRequest("GET", "/health", nil, nil)
}