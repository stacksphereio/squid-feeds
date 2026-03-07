package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"squid-feeds/internal/featureflags"
	"squid-feeds/internal/feeds"
	mw "squid-feeds/internal/http/middleware"
	"squid-feeds/internal/logger"
)

func main() {
	// 1) Feature flags init (non-fatal)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := featureflags.Init(ctx, ""); err != nil {
		log.Printf("feature flags init warning: %v", err)
	} else {
		log.Printf("feature flags ready: offline=%v, logLevel=%s, feedsEnabled=%v",
			featureflags.Values().Offline.IsEnabled(nil),
			featureflags.Values().LogLevel.GetValue(nil),
			featureflags.Values().FeedsEnabled.IsEnabled(nil))
	}
	defer featureflags.Shutdown()

	// 2) Initialize levelled logger from flag & watch for flips
	logger.Init(featureflags.Values().LogLevel.GetValue(nil))
	logger.Infof("log level set to %s", logger.GetLevel())

	go func() {
		prev := featureflags.Values().LogLevel.GetValue(nil)
		for {
			time.Sleep(5 * time.Second)
			cur := featureflags.Values().LogLevel.GetValue(nil)
			if cur != prev {
				logger.SetLevel(cur)
				logger.Infof("log level changed to %s", logger.GetLevel())
				prev = cur
			}
		}
	}()

	// 3) Router
	r := mux.NewRouter()

	// 4) Offline kill-switch middleware
	offlineGate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// always allow health checks
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}
			// block all other requests when Offline flag is ON
			if featureflags.Values().Offline.IsEnabled(nil) {
				http.Error(w, "service temporarily offline", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	r.Use(offlineGate)

	// 5) Request logger (skip noisy health endpoints)
	r.Use(mw.LogRequests(mw.WithSkips("/health", "/ready")))

	// 6) Health endpoints
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	r.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		// No database dependency, always ready if service is running
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	}).Methods(http.MethodGet)

	// 7) Inspect current flag values
	r.HandleFunc("/_flags", func(w http.ResponseWriter, _ *http.Request) {
		flags := featureflags.Values()
		resp := map[string]interface{}{
			"offline":               flags.Offline.IsEnabled(nil),
			"logLevel":              flags.LogLevel.GetValue(nil),
			"feedsEnabled":          flags.FeedsEnabled.IsEnabled(nil),
			"feedsWeatherEnabled":   flags.FeedsWeatherEnabled.IsEnabled(nil),
			"feedsNewsEnabled":      flags.FeedsNewsEnabled.IsEnabled(nil),
			"feedsRegionEUEnabled":  flags.FeedsRegionEUEnabled.IsEnabled(nil),
			"feedsRegionNAEnabled":  flags.FeedsRegionNAEnabled.IsEnabled(nil),
			"feedsRegionSAEnabled":  flags.FeedsRegionSAEnabled.IsEnabled(nil),
			"feedsRegionAsiaEnabled": flags.FeedsRegionAsiaEnabled.IsEnabled(nil),
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods(http.MethodGet)

	// 8) Feed endpoints
	feedHandler := feeds.NewHandler()
	r.HandleFunc("/feeds", feedHandler.HandleFeeds).Methods(http.MethodGet)

	s := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Infof("squid-feeds listening on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}
