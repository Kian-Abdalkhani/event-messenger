package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"event-messenger.com/models"
)

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Remove accents and special characters
	slug = removeAccents(slug)

	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove any characters that aren't alphanumeric or hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Check if slug exists already in db
	baseSlug := slug
	counter := 2
	for {
		_, err := models.GetEventBySlug(slug)
		if err != nil {
			break
		}
		// Slug exists, increment
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	return slug
}

func removeAccents(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if r < unicode.MaxASCII {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
