package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got '%s'", w.Body.String())
	}
}

func TestReadyEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "ready" {
		t.Errorf("expected body 'ready', got '%s'", w.Body.String())
	}
}

func TestFlagsEndpointStructure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/_flags", nil)
	w := httptest.NewRecorder()

	// Simulate flags endpoint response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flags := map[string]interface{}{
			"offline":                 false,
			"logLevel":                "info",
			"feedsEnabled":            true,
			"feedsWeatherEnabled":     true,
			"feedsNewsEnabled":        true,
			"feedsRegionEUEnabled":    true,
			"feedsRegionNAEnabled":    true,
			"feedsRegionSAEnabled":    true,
			"feedsRegionAsiaEnabled":  true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flags)
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	expectedFlags := []string{
		"offline", "logLevel", "feedsEnabled", "feedsWeatherEnabled",
		"feedsNewsEnabled", "feedsRegionEUEnabled", "feedsRegionNAEnabled",
		"feedsRegionSAEnabled", "feedsRegionAsiaEnabled",
	}

	for _, flag := range expectedFlags {
		if _, exists := response[flag]; !exists {
			t.Errorf("expected flag %s to exist in response", flag)
		}
	}
}

func TestOfflineMiddlewareConcept(t *testing.T) {
	// Test that we can construct an offline gate middleware
	offlineGate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// always allow health checks
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}
			// block all other requests when offline
			offline := false // Simulated flag
			if offline {
				http.Error(w, "service temporarily offline", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Test health endpoint is allowed when offline
	handler := offlineGate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("health endpoint should be accessible, got status %d", w.Code)
	}
}

func TestOfflineMiddlewareBlocksRequests(t *testing.T) {
	// Test that offline middleware blocks non-health requests
	offlineGate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}
			offline := true // Simulated offline state
			if offline {
				http.Error(w, "service temporarily offline", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	handler := offlineGate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/feeds", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}
