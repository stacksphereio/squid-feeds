package feeds

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// RegionalClient handles communication with regional feed services
type RegionalClient struct {
	httpClient *http.Client
	baseURLs   map[string]string
}

// NewRegionalClient creates a new regional client with configured URLs
func NewRegionalClient() *RegionalClient {
	return &RegionalClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURLs: map[string]string{
			"EU":   getEnvOrDefault("REEF_EU_URL", "http://reef-eu:8080"),
			"NA":   getEnvOrDefault("REEF_NA_URL", "http://reef-na:8080"),
			"SA":   getEnvOrDefault("REEF_SA_URL", "http://reef-sa:8080"),
			"ASIA": getEnvOrDefault("REEF_ASIA_URL", "http://reef-asia:8080"),
		},
	}
}

// FetchRegionalFeeds calls the regional service for the given region and country
func (c *RegionalClient) FetchRegionalFeeds(region, country string) (*RegionalFeedResponse, error) {
	baseURL, ok := c.baseURLs[region]
	if !ok {
		return nil, fmt.Errorf("unknown region: %s", region)
	}

	url := fmt.Sprintf("%s/regional-feeds?country=%s", baseURL, country)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("regional service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result RegionalFeedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
