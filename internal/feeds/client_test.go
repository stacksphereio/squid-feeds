package feeds

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRegionalClient(t *testing.T) {
	client := NewRegionalClient()

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.httpClient == nil {
		t.Error("expected non-nil http client")
	}

	// Check that all base URLs are configured
	expectedRegions := []string{"EU", "NA", "SA", "ASIA"}
	for _, region := range expectedRegions {
		if _, ok := client.baseURLs[region]; !ok {
			t.Errorf("missing base URL for region %s", region)
		}
	}
}

func TestFetchRegionalFeedsWithMockServer(t *testing.T) {
	// Create a mock regional service
	mockResponse := RegionalFeedResponse{
		Region:  "EU",
		Country: "GB",
		Weather: &WeatherData{
			Summary:      "Clear sky",
			TemperatureC: 20.5,
			FeelsLikeC:   19.0,
		},
		News: []NewsItem{
			{
				ID:          "news-1",
				Title:       "Test News",
				Source:      "Test Source",
				URL:         "https://example.com",
				PublishedAt: "2024-01-01T00:00:00Z",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Query().Get("country") != "GB" {
			t.Errorf("expected country=GB, got %s", r.URL.Query().Get("country"))
		}
		if r.URL.Path != "/regional-feeds" {
			t.Errorf("expected path /regional-feeds, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with test server URL
	client := &RegionalClient{
		httpClient: &http.Client{},
		baseURLs: map[string]string{
			"EU": server.URL,
		},
	}

	// Fetch feeds
	result, err := client.FetchRegionalFeeds("EU", "GB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Region != "EU" {
		t.Errorf("expected region EU, got %s", result.Region)
	}
	if result.Country != "GB" {
		t.Errorf("expected country GB, got %s", result.Country)
	}
	if result.Weather == nil {
		t.Error("expected weather data")
	}
	if len(result.News) != 1 {
		t.Errorf("expected 1 news item, got %d", len(result.News))
	}
}

func TestFetchRegionalFeedsWithUnknownRegion(t *testing.T) {
	client := NewRegionalClient()

	_, err := client.FetchRegionalFeeds("UNKNOWN", "US")
	if err == nil {
		t.Error("expected error for unknown region")
	}
	if err.Error() != "unknown region: UNKNOWN" {
		t.Errorf("expected 'unknown region' error, got: %v", err)
	}
}

func TestFetchRegionalFeedsWithServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := &RegionalClient{
		httpClient: &http.Client{},
		baseURLs: map[string]string{
			"EU": server.URL,
		},
	}

	_, err := client.FetchRegionalFeeds("EU", "GB")
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestFetchRegionalFeedsWithInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := &RegionalClient{
		httpClient: &http.Client{},
		baseURLs: map[string]string{
			"EU": server.URL,
		},
	}

	_, err := client.FetchRegionalFeeds("EU", "GB")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env var set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "env var not set",
			key:          "UNSET_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.key, tt.envValue)
			}

			got := getEnvOrDefault(tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
