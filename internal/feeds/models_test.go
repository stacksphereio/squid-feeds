package feeds

import (
	"encoding/json"
	"testing"
)

func TestWeatherDataSerialization(t *testing.T) {
	weather := WeatherData{
		Summary:      "Partly cloudy",
		TemperatureC: 18.5,
		FeelsLikeC:   17.0,
	}

	data, err := json.Marshal(weather)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded WeatherData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Summary != weather.Summary {
		t.Errorf("expected Summary %s, got %s", weather.Summary, decoded.Summary)
	}
	if decoded.TemperatureC != weather.TemperatureC {
		t.Errorf("expected TemperatureC %f, got %f", weather.TemperatureC, decoded.TemperatureC)
	}
}

func TestNewsItemSerialization(t *testing.T) {
	news := NewsItem{
		ID:          "test-123",
		Title:       "Breaking News",
		Source:      "News Corp",
		URL:         "https://example.com/news",
		PublishedAt: "2024-01-01T12:00:00Z",
	}

	data, err := json.Marshal(news)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded NewsItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != news.ID {
		t.Errorf("expected ID %s, got %s", news.ID, decoded.ID)
	}
	if decoded.Title != news.Title {
		t.Errorf("expected Title %s, got %s", news.Title, decoded.Title)
	}
}

func TestRegionalFeedResponseSerialization(t *testing.T) {
	response := RegionalFeedResponse{
		Region:  "EU",
		Country: "GB",
		Weather: &WeatherData{
			Summary:      "Rainy",
			TemperatureC: 15.0,
			FeelsLikeC:   13.5,
		},
		News: []NewsItem{
			{
				ID:          "1",
				Title:       "News 1",
				Source:      "Source 1",
				URL:         "https://example.com/1",
				PublishedAt: "2024-01-01T00:00:00Z",
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RegionalFeedResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Region != response.Region {
		t.Errorf("expected Region %s, got %s", response.Region, decoded.Region)
	}
	if decoded.Country != response.Country {
		t.Errorf("expected Country %s, got %s", response.Country, decoded.Country)
	}
	if decoded.Weather == nil {
		t.Error("expected weather data")
	}
	if len(decoded.News) != 1 {
		t.Errorf("expected 1 news item, got %d", len(decoded.News))
	}
}

func TestFeedResponseSerialization(t *testing.T) {
	response := FeedResponse{
		User: UserContext{
			ID:      "user123",
			Country: "US",
			Region:  "NA",
		},
		Feeds: Feeds{
			Weather: &WeatherFeed{
				Enabled: true,
				Data: &WeatherData{
					Summary:      "Clear",
					TemperatureC: 20.0,
				},
			},
			News: &NewsFeed{
				Enabled: true,
				Items: []NewsItem{
					{
						ID:    "1",
						Title: "Test News",
					},
				},
			},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded FeedResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.User.ID != response.User.ID {
		t.Errorf("expected User ID %s, got %s", response.User.ID, decoded.User.ID)
	}
	if decoded.Feeds.Weather == nil {
		t.Error("expected weather feed")
	}
	if decoded.Feeds.News == nil {
		t.Error("expected news feed")
	}
}

func TestWeatherDataOmitEmpty(t *testing.T) {
	// Test that weather field is omitted when nil
	response := RegionalFeedResponse{
		Region:  "EU",
		Country: "GB",
		Weather: nil,
		News:    []NewsItem{},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, exists := jsonMap["weather"]; exists {
		t.Error("expected weather field to be omitted when nil")
	}
}

func TestNewsOmitEmpty(t *testing.T) {
	// Test that news field is omitted when nil
	response := RegionalFeedResponse{
		Region:  "EU",
		Country: "GB",
		Weather: &WeatherData{Summary: "Clear"},
		News:    nil,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	if _, exists := jsonMap["news"]; exists {
		t.Error("expected news field to be omitted when nil")
	}
}
