package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/news-reader/internal/models"
	"github.com/news-reader/internal/services"
)

type NewsHandler struct {
	newsService *services.NewsService
}

func NewNewsHandler(newsService *services.NewsService) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
	}
}

func (h *NewsHandler) GetNews(c *gin.Context) {
	news := h.newsService.FetchNews()
	filteredNews := h.newsService.FilterNews(news)
	c.JSON(http.StatusOK, filteredNews)
}

func (h *NewsHandler) GetPreferences(c *gin.Context) {
	prefs := h.newsService.GetPreferences()
	c.JSON(http.StatusOK, prefs)
}

func (h *NewsHandler) UpdatePreferences(c *gin.Context) {
	var newPrefs models.UserPreferences
	if err := c.BindJSON(&newPrefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.newsService.UpdatePreferences(newPrefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newPrefs)
}

func (h *NewsHandler) GetTags(c *gin.Context) {
	systemTags, userTags := h.newsService.GetTags()
	c.JSON(http.StatusOK, gin.H{
		"systemTags": systemTags,
		"userTags":   userTags,
	})
}

func (h *NewsHandler) CreateTag(c *gin.Context) {
	var newTag models.Tag
	if err := c.BindJSON(&newTag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.newsService.CreateTag(newTag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tag)
}

func (h *NewsHandler) UpdateNewsTags(c *gin.Context) {
	newsID := c.Param("id")
	var tags []models.Tag
	if err := c.BindJSON(&tags); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.newsService.UpdateNewsTags(newsID, tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// GetTrendingTopicsHandler returns the current trending topics based on recent news
func (h *NewsHandler) GetTrendingTopicsHandler(c *gin.Context) {
	// Get recent news items
	news := h.newsService.GetAllNews()

	// Get trending topics
	topics := h.newsService.GetTrendingTopics(news)

	// Return trending topics
	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
		"count":  len(topics),
		"time":   time.Now().UTC(),
	})
}

// GetVersionHandler returns the current version information
func (h *NewsHandler) GetVersionHandler(c *gin.Context) {
	// Get version information from main package
	c.JSON(http.StatusOK, gin.H{
		"version":    "0.1.0",
		"buildTime":  time.Now().Format(time.RFC3339),
		"gitCommit":  "development",
	})
}
