package feeds

import (
	"encoding/json"
	"net/http"

	"squid-feeds/internal/auth"
	"squid-feeds/internal/featureflags"
	"squid-feeds/internal/logger"
)

// Handler handles feed requests
type Handler struct {
	client *RegionalClient
}

// NewHandler creates a new feed handler
func NewHandler() *Handler {
	return &Handler{
		client: NewRegionalClient(),
	}
}

// HandleFeeds processes the /feeds endpoint
func (h *Handler) HandleFeeds(w http.ResponseWriter, r *http.Request) {
	// 1. Extract and validate JWT
	tokenStr := auth.GetBearerToken(r)
	if tokenStr == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseToken(tokenStr)
	if err != nil {
		logger.Warnf("JWT validation failed: %v", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.Subject
	if userID == "" {
		http.Error(w, "missing user ID in token", http.StatusBadRequest)
		return
	}

	// 2. Get country from claims (if present)
	country := claims.Country
	if country == "" {
		// Could also look up from user profile service, but for now we'll just use default or return error
		logger.Warnf("no country in JWT for user %s", userID)
		http.Error(w, "user country not available", http.StatusBadRequest)
		return
	}

	// 3. Map country to region
	region := MapCountryToRegion(country)
	if region == "" {
		logger.Warnf("country %s not mapped to any region", country)
		http.Error(w, "unsupported country", http.StatusBadRequest)
		return
	}

	logger.Debugf("user %s from country %s -> region %s", userID, country, region)

	// 4. Evaluate feature flags
	flags := featureflags.Values()

	// Check global master switch
	if !flags.FeedsEnabled.IsEnabled(nil) {
		logger.Debugf("feeds globally disabled for user %s", userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeedResponse{
			User:  UserContext{ID: userID, Country: country, Region: region},
			Feeds: Feeds{},
		})
		return
	}

	// Check regional flag
	regionalEnabled := isRegionEnabled(region, flags)
	if !regionalEnabled {
		logger.Debugf("feeds disabled for region %s", region)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeedResponse{
			User:  UserContext{ID: userID, Country: country, Region: region},
			Feeds: Feeds{},
		})
		return
	}

	weatherEnabled := flags.FeedsWeatherEnabled.IsEnabled(nil)
	newsEnabled := flags.FeedsNewsEnabled.IsEnabled(nil)

	if !weatherEnabled && !newsEnabled {
		logger.Debugf("both weather and news disabled")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FeedResponse{
			User:  UserContext{ID: userID, Country: country, Region: region},
			Feeds: Feeds{},
		})
		return
	}

	// 5. Fetch from regional service
	logger.Infof("fetching feeds for user %s from region %s", userID, region)
	regionalData, err := h.client.FetchRegionalFeeds(region, country)
	if err != nil {
		logger.Errorf("failed to fetch regional feeds: %v", err)
		// Return partial data or error
		http.Error(w, "failed to fetch feeds", http.StatusServiceUnavailable)
		return
	}

	// 6. Compose response based on feature flags
	response := FeedResponse{
		User:  UserContext{ID: userID, Country: country, Region: region},
		Feeds: Feeds{},
	}

	if weatherEnabled && regionalData.Weather != nil {
		response.Feeds.Weather = &WeatherFeed{
			Enabled: true,
			Data:    regionalData.Weather,
		}
	}

	if newsEnabled && regionalData.News != nil {
		response.Feeds.News = &NewsFeed{
			Enabled: true,
			Items:   regionalData.News,
		}
	}

	// 7. Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("failed to encode response: %v", err)
	}
}

// isRegionEnabled checks if feeds are enabled for a specific region
func isRegionEnabled(region string, flags *featureflags.Flags) bool {
	switch region {
	case "EU":
		return flags.FeedsRegionEUEnabled.IsEnabled(nil)
	case "NA":
		return flags.FeedsRegionNAEnabled.IsEnabled(nil)
	case "SA":
		return flags.FeedsRegionSAEnabled.IsEnabled(nil)
	case "ASIA":
		return flags.FeedsRegionAsiaEnabled.IsEnabled(nil)
	default:
		return false
	}
}
