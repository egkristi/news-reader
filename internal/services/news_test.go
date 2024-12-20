package services

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/news-reader/internal/models"
)

func TestNewsService(t *testing.T) {
	// Create a temporary preferences file with initial content
	tmpFile, err := os.CreateTemp("", "preferences-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Create initial preferences
	initialPrefs := models.Preferences{
		Sources: models.DefaultSources,
		Tags:    []models.Tag{},
	}

	// Write initial preferences to file
	prefsData, err := json.Marshal(initialPrefs)
	if err != nil {
		t.Fatalf("Failed to marshal preferences: %v", err)
	}
	if err := os.WriteFile(tmpFile.Name(), prefsData, 0644); err != nil {
		t.Fatalf("Failed to write preferences file: %v", err)
	}

	// Initialize service
	service, err := NewNewsService(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create news service: %v", err)
	}

	// Test preferences initialization
	prefs := service.GetPreferences()
	if prefs == nil {
		t.Fatal("Expected non-nil preferences")
	}
	if len(prefs.Sources) == 0 {
		t.Error("Expected default sources to be present")
	}

	// Test news ID generation
	testItem := models.NewsItem{
		Title:  "Test News",
		Link:   "https://example.com/news/1",
		Source: "Test Source",
	}
	id := service.generateNewsID(testItem)
	if id == "" {
		t.Error("Expected non-empty news ID")
	}

	// Test tag creation
	tag := models.Tag{
		Name:     "Test Tag",
		Color:    "#FF0000",
		Category: "user",
	}
	createdTag, err := service.CreateTag(tag)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}
	if createdTag.ID == "" {
		t.Error("Expected non-empty tag ID")
	}

	// Test news filtering
	items := []models.NewsItem{
		{
			Title:       "Tech News",
			Description: "A technology article",
			Category:    "Technology",
			ContentType: models.TypeRSS,
			Published:   time.Now(),
		},
		{
			Title:       "World News",
			Description: "A world news article",
			Category:    "World News",
			ContentType: models.TypeRSS,
			Published:   time.Now(),
		},
	}

	// Test filtering by category
	service.preferences.Categories = []string{"Technology"}
	filtered := service.FilterNews(items)
	if len(filtered) != 1 || filtered[0].Category != "Technology" {
		t.Error("Expected only technology news to be present")
	}

	// Test filtering by interests
	service.preferences.Categories = nil
	service.preferences.Interests = []string{"world"}
	filtered = service.FilterNews(items)
	if len(filtered) != 1 || filtered[0].Title != "World News" {
		t.Error("Expected only world news to be present")
	}
}

func TestGetTrendingTopics(t *testing.T) {
	service := &NewsService{}

	// Create test news items
	items := []models.NewsItem{
		{
			Title:       "Technology advances in artificial intelligence",
			Description: "New developments in artificial intelligence and machine learning",
		},
		{
			Title:       "AI transforms healthcare industry",
			Description: "Artificial intelligence making breakthroughs in healthcare",
		},
		{
			Title:       "Weather report for today",
			Description: "Sunny weather expected throughout the week",
		},
	}

	// Get trending topics
	topics := service.GetTrendingTopics(items)

	// Verify results
	if len(topics) == 0 {
		t.Error("Expected non-empty trending topics")
	}

	// "artificial" and "intelligence" should be among top topics
	foundAI := false
	for _, topic := range topics {
		if topic.Topic == "artificial" || topic.Topic == "intelligence" {
			if topic.Frequency < 2 {
				t.Errorf("Expected frequency of '%s' to be at least 2, got %d", topic.Topic, topic.Frequency)
			}
			foundAI = true
		}
	}

	if !foundAI {
		t.Error("Expected 'artificial' or 'intelligence' to be among trending topics")
	}
}

func TestLanguageDetection(t *testing.T) {
	service := &NewsService{}

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "English text",
			text:     "The quick brown fox jumps over the lazy dog",
			expected: "english",
		},
		{
			name:     "Spanish text",
			text:     "El perro corre por el parque",
			expected: "spanish",
		},
		{
			name:     "Empty text",
			text:     "",
			expected: "english", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectLanguage(tt.text)
			if result != tt.expected {
				t.Errorf("detectLanguage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRegionDetection(t *testing.T) {
	service := &NewsService{}

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "North American news",
			text:     "Breaking news from the United States",
			expected: "north-america",
		},
		{
			name:     "European news",
			text:     "Latest developments in the European Union",
			expected: "europe",
		},
		{
			name:     "No region",
			text:     "General news without region",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.detectRegion(tt.text)
			if result != tt.expected {
				t.Errorf("detectRegion() = %v, want %v", result, tt.expected)
			}
		})
	}
}
