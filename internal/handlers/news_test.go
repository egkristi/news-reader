package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/news-reader/internal/models"
	"github.com/news-reader/internal/services"
)

func TestGetTrendingTopicsHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Create a temporary preferences file for testing
	tmpFile := t.TempDir() + "/prefs.json"
	service, err := services.NewNewsService(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create news service: %v", err)
	}
	
	handler := NewNewsHandler(service)
	r.GET("/api/news/trending", handler.GetTrendingTopicsHandler)

	// Create test request
	req, err := http.NewRequest(http.MethodGet, "/api/news/trending", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Parse response
	var response struct {
		Topics []services.TrendingTopic `json:"topics"`
		Count  int                      `json:"count"`
		Time   string                   `json:"time"`
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Topics == nil {
		t.Error("Expected non-nil topics array")
	}

	if response.Count != len(response.Topics) {
		t.Errorf("Count mismatch: got %d, want %d", response.Count, len(response.Topics))
	}

	if response.Time == "" {
		t.Error("Expected non-empty timestamp")
	}
}
