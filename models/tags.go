package models

type Tag struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Color    string   `json:"color"`
	Category string   `json:"category"` // system or user
}

type NewsTag struct {
	NewsID string `json:"newsId"`
	TagID  string `json:"tagId"`
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
