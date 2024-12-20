package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
	"github.com/news-reader/config"
	"github.com/news-reader/models"
)

type NewsItem struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Link        string           `json:"link"`
	Description string           `json:"description"`
	Published   time.Time        `json:"published"`
	Source      string           `json:"source"`
	Category    string           `json:"category"`
	ContentType config.ContentType `json:"contentType"`
	Thumbnail   string           `json:"thumbnail,omitempty"`
	Duration    string           `json:"duration,omitempty"`
	AudioURL    string           `json:"audioUrl,omitempty"`
	VideoURL    string           `json:"videoUrl,omitempty"`
	Tags        []models.Tag     `json:"tags"`
	Region      string           `json:"region,omitempty"`
	Language    string           `json:"language,omitempty"`
}

type UserPreferences struct {
	Sources      []config.NewsSource `json:"sources"`
	Interests    []string           `json:"interests"`
	Categories   []string           `json:"categories"`
	ContentTypes []string           `json:"contentTypes"`
	APIKeys      map[string]string  `json:"apiKeys"`
	Tags         []models.Tag       `json:"tags"`
	NewsTags     []models.NewsTag   `json:"newsTags"`
}

var (
	userPrefs UserPreferences
	prefsFile = "preferences.json"
)

// YouTube specific structures
type YouTubeFeed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []struct {
		Title     string `xml:"title"`
		Link      string `xml:"link"`
		Published string `xml:"published"`
		MediaGroup struct {
			Description string `xml:"description"`
			Thumbnail   struct {
				URL string `xml:"url,attr"`
			} `xml:"thumbnail"`
		} `xml:"group"`
	} `xml:"entry"`
}

func loadPreferences() error {
	if _, err := os.Stat(prefsFile); os.IsNotExist(err) {
		userPrefs = UserPreferences{
			Sources:      config.DefaultSources,
			Interests:    []string{},
			Categories:   []string{"General"},
			ContentTypes: []string{"article", "video", "podcast"},
			APIKeys:      make(map[string]string),
			Tags:         []models.Tag{},
			NewsTags:     []models.NewsTag{},
		}
		return savePreferences()
	}

	data, err := os.ReadFile(prefsFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &userPrefs)
}

func savePreferences() error {
	data, err := json.MarshalIndent(userPrefs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(prefsFile, data, 0644)
}

func generateNewsID(item NewsItem) string {
	hash := sha256.New()
	hash.Write([]byte(item.Title + item.Link + item.Source))
	return hex.EncodeToString(hash.Sum(nil))
}

func detectRegion(text string) string {
	regions := map[string][]string{
		"north-america": {"USA", "Canada", "Mexico", "United States", "American"},
		"south-america": {"Brazil", "Argentina", "Chile", "Colombia", "Venezuela"},
		"europe":        {"EU", "European Union", "UK", "Britain", "Germany", "France", "Italy", "Spain"},
		"asia":          {"China", "Japan", "India", "Korea", "Asian"},
		"africa":        {"Africa", "Nigeria", "Egypt", "South Africa", "Kenya"},
		"oceania":       {"Australia", "New Zealand", "Pacific"},
	}

	text = strings.ToLower(text)
	for region, keywords := range regions {
		for _, keyword := range keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				return region
			}
		}
	}
	return ""
}

func detectLanguage(text string) string {
	// Simple language detection based on common words
	languages := map[string][]string{
		"english": {"the", "and", "in", "of", "to"},
		"spanish": {"el", "la", "en", "de", "por"},
		"french":  {"le", "la", "les", "en", "de"},
		"german":  {"der", "die", "das", "und", "in"},
	}

	wordCounts := make(map[string]int)
	words := strings.Fields(strings.ToLower(text))

	for _, word := range words {
		for lang, markers := range languages {
			for _, marker := range markers {
				if word == marker {
					wordCounts[lang]++
				}
			}
		}
	}

	maxCount := 0
	detectedLang := "english" // default
	for lang, count := range wordCounts {
		if count > maxCount {
			maxCount = count
			detectedLang = lang
		}
	}

	return detectedLang
}

func autoTagNews(item *NewsItem) {
	// Auto-detect region and language
	combinedText := item.Title + " " + item.Description
	item.Region = detectRegion(combinedText)
	item.Language = detectLanguage(combinedText)

	// Initialize tags slice
	item.Tags = []models.Tag{}

	// Add region tag if detected
	if item.Region != "" {
		for _, tag := range models.DefaultTags {
			if tag.ID == item.Region && tag.Category == "region" {
				item.Tags = append(item.Tags, tag)
				break
			}
		}
	}

	// Add language tag
	for _, tag := range models.DefaultTags {
		if tag.ID == item.Language && tag.Category == "language" {
			item.Tags = append(item.Tags, tag)
			break
		}
	}

	// Add topic tags based on content analysis
	topicKeywords := map[string][]string{
		"politics":      {"politics", "government", "election", "president", "minister"},
		"economy":       {"economy", "market", "stock", "trade", "financial"},
		"technology":    {"technology", "software", "digital", "cyber", "AI"},
		"science":       {"science", "research", "study", "discovery"},
		"health":        {"health", "medical", "disease", "treatment", "covid"},
		"sports":        {"sports", "game", "tournament", "championship", "player"},
		"entertainment": {"entertainment", "movie", "music", "celebrity", "art"},
		"environment":   {"environment", "climate", "pollution", "sustainable"},
	}

	text := strings.ToLower(combinedText)
	for topic, keywords := range topicKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				for _, tag := range models.DefaultTags {
					if tag.ID == topic && tag.Category == "topic" {
						item.Tags = append(item.Tags, tag)
						break
					}
				}
				break
			}
		}
	}

	// Add user-defined tags if they match any criteria
	for _, userTag := range userPrefs.Tags {
		if strings.Contains(strings.ToLower(combinedText), strings.ToLower(userTag.Name)) {
			item.Tags = append(item.Tags, userTag)
		}
	}
}

func fetchNewsFromSource(src config.NewsSource) ([]NewsItem, error) {
	items, err := fetchRawNewsFromSource(src)
	if err != nil {
		return nil, err
	}

	// Process each item to add IDs and tags
	for i := range items {
		items[i].ID = generateNewsID(items[i])
		autoTagNews(&items[i])
	}

	return items, nil
}

func fetchRawNewsFromSource(src config.NewsSource) ([]NewsItem, error) {
	switch src.ContentType {
	case config.TypeRSS:
		return fetchRSSFeed(src)
	case config.TypeVideo:
		return fetchYouTubeFeed(src)
	case config.TypePodcast:
		return fetchPodcastFeed(src)
	case config.TypeAPI:
		return fetchAPIContent(src)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", src.ContentType)
	}
}

func fetchRSSFeed(src config.NewsSource) ([]NewsItem, error) {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
	
	feed, err := parser.ParseURL(src.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing RSS feed from %s: %v", src.Name, err)
	}

	if feed == nil {
		return nil, fmt.Errorf("received nil feed from %s", src.Name)
	}

	var items []NewsItem
	for _, item := range feed.Items {
		if item == nil {
			continue
		}

		published := time.Now()
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}

		description := item.Description
		if description == "" && item.Content != "" {
			description = item.Content
		}

		newsItem := NewsItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: description,
			Published:   published,
			Source:      src.Name,
			Category:    src.Category,
			ContentType: src.ContentType,
		}

		// Try to extract image from content if available
		if item.Image != nil && item.Image.URL != "" {
			newsItem.Thumbnail = item.Image.URL
		} else if item.ITunesExt != nil && item.ITunesExt.Image != "" {
			newsItem.Thumbnail = item.ITunesExt.Image
		}

		items = append(items, newsItem)
	}

	if len(items) == 0 {
		log.Printf("Warning: No items found in feed from %s", src.Name)
	}

	return items, nil
}

func fetchYouTubeFeed(src config.NewsSource) ([]NewsItem, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Get(src.URL)
	if err != nil {
		return nil, fmt.Errorf("error fetching YouTube feed from %s: %v", src.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d from %s", resp.StatusCode, src.Name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body from %s: %v", src.Name, err)
	}

	var feed YouTubeFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("error parsing YouTube feed from %s: %v", src.Name, err)
	}

	var items []NewsItem
	for _, entry := range feed.Entries {
		published, err := time.Parse(time.RFC3339, entry.Published)
		if err != nil {
			published = time.Now()
		}

		if entry.Title == "" || entry.Link == "" {
			continue
		}

		items = append(items, NewsItem{
			Title:       entry.Title,
			Link:        entry.Link,
			Description: entry.MediaGroup.Description,
			Published:   published,
			Source:      src.Name,
			Category:    src.Category,
			ContentType: src.ContentType,
			Thumbnail:   entry.MediaGroup.Thumbnail.URL,
			VideoURL:    entry.Link,
		})
	}

	if len(items) == 0 {
		log.Printf("Warning: No items found in YouTube feed from %s", src.Name)
	}

	return items, nil
}

func fetchPodcastFeed(src config.NewsSource) ([]NewsItem, error) {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
	
	feed, err := parser.ParseURL(src.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing podcast feed from %s: %v", src.Name, err)
	}

	if feed == nil {
		return nil, fmt.Errorf("received nil feed from %s", src.Name)
	}

	var items []NewsItem
	for _, item := range feed.Items {
		if item == nil {
			continue
		}

		published := time.Now()
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}

		description := item.Description
		if description == "" && item.Content != "" {
			description = item.Content
		}

		newsItem := NewsItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: description,
			Published:   published,
			Source:      src.Name,
			Category:    src.Category,
			ContentType: src.ContentType,
		}

		// Extract audio URL from enclosures if available
		if len(item.Enclosures) > 0 && item.Enclosures[0].URL != "" {
			newsItem.AudioURL = item.Enclosures[0].URL
		}

		// Try to extract duration if available
		if item.ITunesExt != nil {
			newsItem.Duration = item.ITunesExt.Duration
			if newsItem.Thumbnail == "" {
				newsItem.Thumbnail = item.ITunesExt.Image
			}
		}

		items = append(items, newsItem)
	}

	if len(items) == 0 {
		log.Printf("Warning: No items found in podcast feed from %s", src.Name)
	}

	return items, nil
}

func fetchAPIContent(src config.NewsSource) ([]NewsItem, error) {
	// Add API key to request if available
	apiKey := userPrefs.APIKeys[src.Name]
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found for %s", src.Name)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	req, err := http.NewRequest("GET", src.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %v", src.Name, err)
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to %s: %v", src.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code %d from %s", resp.StatusCode, src.Name)
	}

	// Parse response based on the API source
	var result struct {
		Articles []struct {
			Title       string    `json:"title"`
			URL         string    `json:"url"`
			Description string    `json:"description"`
			PublishedAt time.Time `json:"publishedAt"`
			URLToImage  string    `json:"urlToImage"`
		} `json:"articles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response from %s: %v", src.Name, err)
	}

	var items []NewsItem
	for _, article := range result.Articles {
		if article.Title == "" || article.URL == "" {
			continue
		}

		items = append(items, NewsItem{
			Title:       article.Title,
			Link:        article.URL,
			Description: article.Description,
			Published:   article.PublishedAt,
			Source:      src.Name,
			Category:    src.Category,
			ContentType: src.ContentType,
			Thumbnail:   article.URLToImage,
		})
	}

	if len(items) == 0 {
		log.Printf("Warning: No items found in API response from %s", src.Name)
	}

	return items, nil
}

func fetchNews(sources []config.NewsSource) []NewsItem {
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		newsItems []NewsItem
	)

	// Create a channel to collect errors
	errorCh := make(chan error, len(sources))

	for _, source := range sources {
		if !source.Enabled {
			continue
		}

		wg.Add(1)
		go func(src config.NewsSource) {
			defer wg.Done()

			items, err := fetchNewsFromSource(src)
			if err != nil {
				log.Printf("Error fetching from %s: %v", src.Name, err)
				errorCh <- err
				return
			}

			mu.Lock()
			newsItems = append(newsItems, items...)
			mu.Unlock()
		}(source)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errorCh)

	// Log all errors
	for err := range errorCh {
		log.Printf("Fetch error: %v", err)
	}

	// Sort news items by published date (newest first)
	sort.Slice(newsItems, func(i, j int) bool {
		return newsItems[i].Published.After(newsItems[j].Published)
	})

	return newsItems
}

func filterNews(items []NewsItem, prefs UserPreferences) []NewsItem {
	if len(prefs.Interests) == 0 && len(prefs.Categories) == 0 && len(prefs.ContentTypes) == 0 {
		return items
	}

	var filtered []NewsItem
	for _, item := range items {
		// Check categories
		categoryMatch := len(prefs.Categories) == 0
		for _, cat := range prefs.Categories {
			if strings.EqualFold(item.Category, cat) {
				categoryMatch = true
				break
			}
		}

		// Check content types
		contentTypeMatch := len(prefs.ContentTypes) == 0
		for _, ct := range prefs.ContentTypes {
			if string(item.ContentType) == ct {
				contentTypeMatch = true
				break
			}
		}

		// Check interests
		interestMatch := len(prefs.Interests) == 0
		for _, interest := range prefs.Interests {
			if strings.Contains(strings.ToLower(item.Title), strings.ToLower(interest)) ||
				strings.Contains(strings.ToLower(item.Description), strings.ToLower(interest)) {
				interestMatch = true
				break
			}
		}

		if categoryMatch && interestMatch && contentTypeMatch {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func main() {
	if err := loadPreferences(); err != nil {
		log.Fatal("Error loading preferences:", err)
	}

	r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/api/news", func(c *gin.Context) {
		news := fetchNews(userPrefs.Sources)
		filteredNews := filterNews(news, userPrefs)
		c.JSON(http.StatusOK, filteredNews)
	})

	r.GET("/api/preferences", func(c *gin.Context) {
		c.JSON(http.StatusOK, userPrefs)
	})

	r.POST("/api/preferences", func(c *gin.Context) {
		var newPrefs UserPreferences
		if err := c.BindJSON(&newPrefs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userPrefs = newPrefs
		if err := savePreferences(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, userPrefs)
	})

	// Tag management endpoints
	r.GET("/api/tags", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"systemTags": models.DefaultTags,
			"userTags":   userPrefs.Tags,
		})
	})

	r.POST("/api/tags", func(c *gin.Context) {
		var newTag models.Tag
		if err := c.BindJSON(&newTag); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate a unique ID for the tag
		hash := sha256.New()
		hash.Write([]byte(newTag.Name + time.Now().String()))
		newTag.ID = hex.EncodeToString(hash.Sum(nil))[:8]
		newTag.Category = "user"

		userPrefs.Tags = append(userPrefs.Tags, newTag)
		if err := savePreferences(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, newTag)
	})

	r.POST("/api/news/:id/tags", func(c *gin.Context) {
		newsID := c.Param("id")
		var tags []models.Tag
		if err := c.BindJSON(&tags); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update news tags
		var newTags []models.NewsTag
		for _, tag := range tags {
			newTags = append(newTags, models.NewsTag{
				NewsID: newsID,
				TagID:  tag.ID,
			})
		}

		// Remove old tags for this news item
		existingTags := []models.NewsTag{}
		for _, nt := range userPrefs.NewsTags {
			if nt.NewsID != newsID {
				existingTags = append(existingTags, nt)
			}
		}
		userPrefs.NewsTags = append(existingTags, newTags...)

		if err := savePreferences(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, tags)
	})

	log.Fatal(r.Run(":8082"))
}
