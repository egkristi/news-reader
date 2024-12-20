package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/news-reader/internal/handlers"
	"github.com/news-reader/internal/services"
)

var (
	// Version information
	Version   = "0.1.0"
	BuildTime = ""
	GitCommit = ""
)

func main() {
	// Command line flags
	var (
		port      = flag.Int("port", 8082, "Server port")
		prefsFile = flag.String("prefs", "preferences.json", "Path to preferences file")
		debug     = flag.Bool("debug", true, "Enable debug mode")
	)
	flag.Parse()

	// Set Gin mode
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Ensure preferences directory exists
	prefsDir := filepath.Dir(*prefsFile)
	if err := os.MkdirAll(prefsDir, 0755); err != nil {
		log.Fatalf("Failed to create preferences directory: %v", err)
	}

	// Initialize services
	newsService, err := services.NewNewsService(*prefsFile)
	if err != nil {
		log.Fatalf("Failed to initialize news service: %v", err)
	}

	// Initialize handlers
	newsHandler := handlers.NewNewsHandler(newsService)

	// Setup router
	r := gin.Default()

	// Setup routes
	setupRoutes(r, newsHandler)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine, newsHandler *handlers.NewsHandler) {
	// API routes
	api := r.Group("/api")
	{
		api.GET("/news", newsHandler.GetNews)
		api.GET("/news/trending", newsHandler.GetTrendingTopicsHandler)
		api.GET("/version", newsHandler.GetVersionHandler)
		api.GET("/tags", newsHandler.GetTags)
		api.POST("/tags", newsHandler.CreateTag)
		api.PUT("/preferences", newsHandler.UpdatePreferences)
		api.GET("/preferences", newsHandler.GetPreferences)
		api.POST("/news/:id/tags", newsHandler.UpdateNewsTags)
	}

	// Serve static files
	r.Static("/static", "web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Serve index page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
}
