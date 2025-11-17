package utils

import (
	"net/http"
	"os"
)

// GetBaseURL returns the base URL from env or request
func GetBaseURL(r *http.Request) string {
	// Try environment variable first (for production)
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL
	}

	// Fall back to request host (for development)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// GetEventURL generates the full URL for an event slug
func GetEventURL(slug string, r *http.Request) string {
	return GetBaseURL(r) + "/event/" + slug + "/messages"
}
