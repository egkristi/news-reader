package config

type ContentType string

const (
	TypeArticle  ContentType = "article"
	TypePodcast  ContentType = "podcast"
	TypeVideo    ContentType = "video"
	TypeRSS      ContentType = "rss"
	TypeAPI      ContentType = "api"
)

type NewsSource struct {
	Name        string      `json:"name"`
	URL         string      `json:"url"`
	Category    string      `json:"category"`
	ContentType ContentType `json:"contentType"`
	APIKey      string      `json:"apiKey,omitempty"`
	Enabled     bool        `json:"enabled"`
}

var DefaultSources = []NewsSource{
	// General News - Free RSS Sources
	{"Reuters World News", "https://feeds.reuters.com/reuters/worldNews", "General", TypeRSS, "", true},
	{"Associated Press", "https://feeds.apnews.com/apnews", "General", TypeRSS, "", true},
	{"NPR News", "https://feeds.npr.org/1001/rss.xml", "General", TypeRSS, "", true},
	{"Voice of America", "https://www.voanews.com/api/zyq$metyqy", "General", TypeRSS, "", true},
	
	// Technology News - Free Sources
	{"TechCrunch", "https://techcrunch.com/feed/", "Technology", TypeRSS, "", true},
	{"Ars Technica", "https://feeds.arstechnica.com/arstechnica/index", "Technology", TypeRSS, "", true},
	{"The Verge", "https://www.theverge.com/rss/index.xml", "Technology", TypeRSS, "", true},
	{"Engadget", "https://www.engadget.com/rss.xml", "Technology", TypeRSS, "", true},
	{"Hacker News", "https://news.ycombinator.com/rss", "Technology", TypeRSS, "", true},
	
	// Science News - Free Sources
	{"Science Daily", "https://www.sciencedaily.com/rss/all.xml", "Science", TypeRSS, "", true},
	{"Live Science", "https://www.livescience.com/feeds/all", "Science", TypeRSS, "", true},
	{"Space.com", "https://www.space.com/feeds/all", "Science", TypeRSS, "", true},
	
	// Video Content - Free Sources
	{"NASA Video Updates", "https://www.nasa.gov/rss/dyn/video_collection.rss", "Science", TypeVideo, "", true},
	{"TED Talks", "https://feeds.feedburner.com/tedtalks_video", "Education", TypeVideo, "", true},
	
	// Podcasts - Free Sources
	{"NPR Technology Podcast", "https://feeds.npr.org/510051/podcast.xml", "Technology", TypePodcast, "", true},
	{"Science Friday", "https://feeds.sciencefriday.com/scifri", "Science", TypePodcast, "", true},
	{"TED Radio Hour", "https://feeds.npr.org/510298/podcast.xml", "Education", TypePodcast, "", true},
	
	// Entertainment - Free Sources
	{"IGN", "https://feeds.ign.com/ign/all", "Entertainment", TypeRSS, "", true},
	{"Polygon", "https://www.polygon.com/rss/index.xml", "Entertainment", TypeRSS, "", true},
	
	// Reddit Feeds - Free Sources
	{"Reddit Technology", "https://www.reddit.com/r/technology/.rss", "Technology", TypeRSS, "", true},
	{"Reddit Science", "https://www.reddit.com/r/science/.rss", "Science", TypeRSS, "", true},
	{"Reddit World News", "https://www.reddit.com/r/worldnews/.rss", "General", TypeRSS, "", true},
}
