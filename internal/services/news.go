package services

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

	"github.com/mmcdole/gofeed"
	"github.com/news-reader/internal/models"
)

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

type NewsService struct {
	preferences *models.UserPreferences
	prefsFile   string
	mu          sync.RWMutex
	newsCache   map[string][]models.NewsItem
}

type TrendingTopic struct {
	Topic     string `json:"topic"`
	Frequency int    `json:"frequency"`
}

func NewNewsService(prefsFile string) (*NewsService, error) {
	service := &NewsService{
		prefsFile: prefsFile,
		newsCache: make(map[string][]models.NewsItem),
	}

	if err := service.loadPreferences(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *NewsService) loadPreferences() error {
	if _, err := os.Stat(s.prefsFile); os.IsNotExist(err) {
		s.preferences = models.NewDefaultPreferences()
		return s.savePreferences()
	}

	data, err := os.ReadFile(s.prefsFile)
	if err != nil {
		return err
	}

	s.preferences = &models.UserPreferences{}
	return json.Unmarshal(data, s.preferences)
}

func (s *NewsService) savePreferences() error {
	data, err := json.MarshalIndent(s.preferences, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.prefsFile, data, 0644)
}

func (s *NewsService) GetPreferences() *models.UserPreferences {
	return s.preferences
}

func (s *NewsService) UpdatePreferences(prefs models.UserPreferences) error {
	s.preferences = &prefs
	return s.savePreferences()
}

func (s *NewsService) GetTags() ([]models.Tag, []models.Tag) {
	return models.DefaultTags, s.preferences.Tags
}

func (s *NewsService) CreateTag(tag models.Tag) (models.Tag, error) {
	// Generate a unique ID for the tag
	hash := sha256.New()
	hash.Write([]byte(tag.Name + time.Now().String()))
	tag.ID = hex.EncodeToString(hash.Sum(nil))[:8]
	tag.Category = "user"

	s.preferences.Tags = append(s.preferences.Tags, tag)
	if err := s.savePreferences(); err != nil {
		return models.Tag{}, err
	}

	return tag, nil
}

func (s *NewsService) UpdateNewsTags(newsID string, tags []models.Tag) error {
	var newTags []models.NewsTag
	for _, tag := range tags {
		newTags = append(newTags, models.NewsTag{
			NewsID: newsID,
			TagID:  tag.ID,
		})
	}

	// Remove old tags for this news item
	existingTags := []models.NewsTag{}
	for _, nt := range s.preferences.NewsTags {
		if nt.NewsID != newsID {
			existingTags = append(existingTags, nt)
		}
	}
	s.preferences.NewsTags = append(existingTags, newTags...)

	return s.savePreferences()
}

func (s *NewsService) generateNewsID(item models.NewsItem) string {
	hash := sha256.New()
	hash.Write([]byte(item.Title + item.Link + item.Source))
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *NewsService) detectRegion(text string) string {
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

func (s *NewsService) detectLanguage(text string) string {
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

func (s *NewsService) autoTagNews(item *models.NewsItem) {
	// Auto-detect region and language
	combinedText := item.Title + " " + item.Description
	item.Region = s.detectRegion(combinedText)
	item.Language = s.detectLanguage(combinedText)

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
	for _, userTag := range s.preferences.Tags {
		if strings.Contains(strings.ToLower(combinedText), strings.ToLower(userTag.Name)) {
			item.Tags = append(item.Tags, userTag)
		}
	}
}

func (s *NewsService) fetchRSSFeed(src models.NewsSource) ([]models.NewsItem, error) {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
		},
	}
	
	// Create a request to set custom headers
	req, err := http.NewRequest("GET", src.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %v", src.Name, err)
	}

	// Set common headers that most RSS feeds expect
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; NewsReader/1.0)")
	req.Header.Set("Accept", "application/rss+xml, application/xml, application/atom+xml, text/xml")

	// Parse the feed with custom request
	feed, err := parser.ParseURLWithContext(src.URL, req.Context())
	if err != nil {
		return nil, fmt.Errorf("error parsing RSS feed from %s: %v", src.Name, err)
	}

	if feed == nil {
		return nil, fmt.Errorf("received nil feed from %s", src.Name)
	}

	var items []models.NewsItem
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

		newsItem := models.NewsItem{
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

func (s *NewsService) fetchYouTubeFeed(src models.NewsSource) ([]models.NewsItem, error) {
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

	var items []models.NewsItem
	for _, entry := range feed.Entries {
		published, err := time.Parse(time.RFC3339, entry.Published)
		if err != nil {
			published = time.Now()
		}

		if entry.Title == "" || entry.Link == "" {
			continue
		}

		items = append(items, models.NewsItem{
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

func (s *NewsService) fetchPodcastFeed(src models.NewsSource) ([]models.NewsItem, error) {
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

	var items []models.NewsItem
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

		newsItem := models.NewsItem{
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

func (s *NewsService) fetchAPIContent(src models.NewsSource) ([]models.NewsItem, error) {
	// Add API key to request if available
	apiKey := s.preferences.APIKeys[src.Name]
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

	var items []models.NewsItem
	for _, article := range result.Articles {
		if article.Title == "" || article.URL == "" {
			continue
		}

		items = append(items, models.NewsItem{
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

func (s *NewsService) fetchRawNewsFromSource(src models.NewsSource) ([]models.NewsItem, error) {
	switch src.ContentType {
	case models.TypeRSS:
		return s.fetchRSSFeed(src)
	case models.TypeVideo:
		return s.fetchYouTubeFeed(src)
	case models.TypePodcast:
		return s.fetchPodcastFeed(src)
	case models.TypeAPI:
		return s.fetchAPIContent(src)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", src.ContentType)
	}
}

func (s *NewsService) fetchNewsFromSource(src models.NewsSource) ([]models.NewsItem, error) {
	items, err := s.fetchRawNewsFromSource(src)
	if err != nil {
		return nil, err
	}

	// Process each item to add IDs and tags
	for i := range items {
		items[i].ID = s.generateNewsID(items[i])
		s.autoTagNews(&items[i])
	}

	return items, nil
}

func (s *NewsService) FetchNews() []models.NewsItem {
	var wg sync.WaitGroup

	// Fetch news from each source
	for _, source := range s.preferences.Sources {
		if !source.Enabled {
			continue
		}

		wg.Add(1)
		go func(src models.NewsSource) {
			defer wg.Done()

			items, err := s.fetchNewsFromSource(src)
			if err != nil {
				log.Printf("Fetch error: %v", err)
				return
			}

			s.mu.Lock()
			s.newsCache[src.Name] = items
			s.mu.Unlock()
		}(source)
	}

	wg.Wait()

	// Return all news items
	return s.GetAllNews()
}

func (s *NewsService) GetAllNews() []models.NewsItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allNews []models.NewsItem
	for _, items := range s.newsCache {
		allNews = append(allNews, items...)
	}
	return allNews
}

func (s *NewsService) FilterNews(items []models.NewsItem) []models.NewsItem {
	if len(s.preferences.Interests) == 0 && len(s.preferences.Categories) == 0 && len(s.preferences.ContentTypes) == 0 {
		return items
	}

	var filtered []models.NewsItem
	for _, item := range items {
		// Check content type
		if len(s.preferences.ContentTypes) > 0 {
			found := false
			for _, ct := range s.preferences.ContentTypes {
				if string(item.ContentType) == ct {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Check category
		if len(s.preferences.Categories) > 0 {
			found := false
			for _, cat := range s.preferences.Categories {
				if item.Category == cat {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Check interests
		if len(s.preferences.Interests) > 0 {
			found := false
			text := strings.ToLower(item.Title + " " + item.Description)
			for _, interest := range s.preferences.Interests {
				if strings.Contains(text, strings.ToLower(interest)) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, item)
	}

	return filtered
}

func (s *NewsService) GetTrendingTopics(items []models.NewsItem) []TrendingTopic {
	topicFrequency := make(map[string]int)
	stopWords := map[string]bool{
		"the": true, "be": true, "to": true, "of": true, "and": true,
		"a": true, "in": true, "that": true, "have": true, "i": true,
		"it": true, "for": true, "not": true, "on": true, "with": true,
		"he": true, "as": true, "you": true, "do": true, "at": true,
	}

	// Analyze each news item
	for _, item := range items {
		// Combine title and description for analysis
		text := strings.ToLower(item.Title + " " + item.Description)
		words := strings.Fields(text)

		// Count word frequencies, excluding stop words
		for _, word := range words {
			// Skip short words and stop words
			if len(word) < 4 || stopWords[word] {
				continue
			}
			topicFrequency[word]++
		}
	}

	// Convert map to slice for sorting
	var topics []TrendingTopic
	for word, freq := range topicFrequency {
		if freq > 1 { // Only include topics that appear more than once
			topics = append(topics, TrendingTopic{
				Topic:     word,
				Frequency: freq,
			})
		}
	}

	// Sort by frequency (highest first)
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Frequency > topics[j].Frequency
	})

	// Return top 10 topics
	if len(topics) > 10 {
		return topics[:10]
	}
	return topics
}
