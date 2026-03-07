package feeds

// WeatherData represents weather information
type WeatherData struct {
	Summary      string  `json:"summary"`
	TemperatureC float64 `json:"temperatureC"`
	FeelsLikeC   float64 `json:"feelsLikeC,omitempty"`
}

// NewsItem represents a single news article
type NewsItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Source      string `json:"source"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"` // ISO datetime string
}

// WeatherFeed represents a weather feed with enabled flag
type WeatherFeed struct {
	Enabled bool         `json:"enabled"`
	Data    *WeatherData `json:"data,omitempty"`
}

// NewsFeed represents a news feed with enabled flag
type NewsFeed struct {
	Enabled bool       `json:"enabled"`
	Items   []NewsItem `json:"items"`
}

// FeedResponse is the top-level response structure
type FeedResponse struct {
	User  UserContext `json:"user"`
	Feeds Feeds       `json:"feeds"`
}

// UserContext provides user information in the response
type UserContext struct {
	ID      string `json:"id"`
	Country string `json:"country"`
	Region  string `json:"region"`
}

// Feeds contains all feed types
type Feeds struct {
	Weather *WeatherFeed `json:"weather,omitempty"`
	News    *NewsFeed    `json:"news,omitempty"`
}

// RegionalFeedResponse represents the response from a regional service
type RegionalFeedResponse struct {
	Region  string       `json:"region"`
	Country string       `json:"country"`
	Weather *WeatherData `json:"weather,omitempty"`
	News    []NewsItem   `json:"news,omitempty"`
}
