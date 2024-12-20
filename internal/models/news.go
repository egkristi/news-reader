package models

import "time"

type ContentType string

const (
	TypeRSS     ContentType = "rss"
	TypeVideo   ContentType = "video"
	TypePodcast ContentType = "podcast"
	TypeAPI     ContentType = "api"
)

type NewsSource struct {
	Name        string      `json:"name"`
	URL         string      `json:"url"`
	Category    string      `json:"category"`
	ContentType ContentType `json:"contentType"`
	Enabled     bool        `json:"enabled"`
}

type NewsItem struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Link        string      `json:"link"`
	Description string      `json:"description"`
	Published   time.Time   `json:"published"`
	Source      string      `json:"source"`
	Category    string      `json:"category"`
	ContentType ContentType `json:"contentType"`
	Thumbnail   string      `json:"thumbnail,omitempty"`
	Duration    string      `json:"duration,omitempty"`
	AudioURL    string      `json:"audioUrl,omitempty"`
	VideoURL    string      `json:"videoUrl,omitempty"`
	Tags        []Tag       `json:"tags"`
	Region      string      `json:"region,omitempty"`
	Language    string      `json:"language,omitempty"`
}

type Tag struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"` // system or user
}

type NewsTag struct {
	NewsID string `json:"newsId"`
	TagID  string `json:"tagId"`
}

type UserPreferences struct {
	Sources      []NewsSource    `json:"sources"`
	Interests    []string        `json:"interests"`
	Categories   []string        `json:"categories"`
	ContentTypes []string        `json:"contentTypes"`
	APIKeys      map[string]string `json:"apiKeys"`
	Tags         []Tag          `json:"tags"`
	NewsTags     []NewsTag      `json:"newsTags"`
}

type Preferences struct {
	Sources    []NewsSource `json:"sources"`
	Tags       []Tag        `json:"tags"`
	Categories []string     `json:"categories"`
	Interests  []string     `json:"interests"`
}

func NewDefaultPreferences() *UserPreferences {
	return &UserPreferences{
		Sources:      DefaultSources,
		Interests:    []string{},
		Categories:   []string{"General"},
		ContentTypes: []string{"article", "video", "podcast"},
		APIKeys:      make(map[string]string),
		Tags:         []Tag{},
		NewsTags:     []NewsTag{},
	}
}

// Default news sources
var DefaultSources = []NewsSource{
	{
		Name:        "NPR News",
		URL:         "https://feeds.npr.org/1001/rss.xml",
		Category:    "General",
		ContentType: TypeRSS,
		Enabled:     true,
	},
	{
		Name:        "BBC World",
		URL:         "http://feeds.bbci.co.uk/news/world/rss.xml",
		Category:    "World News",
		ContentType: TypeRSS,
		Enabled:     true,
	},
	{
		Name:        "The Guardian",
		URL:         "https://www.theguardian.com/world/rss",
		Category:    "World News",
		ContentType: TypeRSS,
		Enabled:     true,
	},
	{
		Name:        "TechCrunch",
		URL:         "https://techcrunch.com/feed/",
		Category:    "Technology",
		ContentType: TypeRSS,
		Enabled:     true,
	},
}

// Predefined system tags
var DefaultTags = []Tag{
	// Regions
	{ID: "north-america", Name: "North America", Color: "#FF6B6B", Category: "region"},
	{ID: "south-america", Name: "South America", Color: "#4ECDC4", Category: "region"},
	{ID: "europe", Name: "Europe", Color: "#45B7D1", Category: "region"},
	{ID: "asia", Name: "Asia", Color: "#96CEB4", Category: "region"},
	{ID: "africa", Name: "Africa", Color: "#FFEEAD", Category: "region"},
	{ID: "oceania", Name: "Oceania", Color: "#D4A5A5", Category: "region"},

	// Languages
	{ID: "english", Name: "English", Color: "#9B59B6", Category: "language"},
	{ID: "spanish", Name: "Spanish", Color: "#E67E22", Category: "language"},
	{ID: "french", Name: "French", Color: "#F1C40F", Category: "language"},
	{ID: "german", Name: "German", Color: "#2ECC71", Category: "language"},

	// Topics
	{ID: "politics", Name: "Politics", Color: "#E74C3C", Category: "topic"},
	{ID: "economy", Name: "Economy", Color: "#27AE60", Category: "topic"},
	{ID: "technology", Name: "Technology", Color: "#3498DB", Category: "topic"},
	{ID: "science", Name: "Science", Color: "#8E44AD", Category: "topic"},
	{ID: "health", Name: "Health", Color: "#2C3E50", Category: "topic"},
	{ID: "sports", Name: "Sports", Color: "#F39C12", Category: "topic"},
	{ID: "entertainment", Name: "Entertainment", Color: "#D35400", Category: "topic"},
	{ID: "environment", Name: "Environment", Color: "#16A085", Category: "topic"},
}
