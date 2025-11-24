package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"event-messenger.com/models"
)

func GetFunnelURL(slug string) (string, error) {
	cmd := exec.Command(
		"docker", "exec", "ts",
		"tailscale", "funnel", "status",
		"--json",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get tailscale status: %w", err)
	}

	// Handling for blank output (no funnel active)
	if string(output) == "{}" {
		return "", nil
	}

	// Parse JSON to get funnel domain
	var status map[string]any
	if err := json.Unmarshal(output, &status); err != nil {
		return "", fmt.Errorf("failed to parse tailscale status: %w", err)
	}

	var baseURL string

	if allowFunnel, ok := status["AllowFunnel"]; ok {
		if funnelMap, ok := allowFunnel.(map[string]any); ok {
			for key, value := range funnelMap {
				// Need to assert the value type as a boolean
				if boolVal, ok := value.(bool); ok && boolVal {
					baseURL = key

					// remove port number if present
					if idx := strings.LastIndex(baseURL, ":"); idx != -1 {
						baseURL = baseURL[:idx]
					}

					// add in /events/ at the end
					baseURL = "https://" + baseURL + "/events/"

				}
			}
		}
	}

	return baseURL + slug, nil

}

// Create tailscale funnel
func CreateFunnel() error {
	cmd := exec.Command(
		"docker", "exec", "ts",
		"tailscale", "funnel",
		"--bg",
		"http://localhost:8080",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to create funnel: - Output: %s", string(output))
		return fmt.Errorf("funnel creation failed: %w", err)
	}

	log.Printf("Successfully created funnel Output: %s", string(output))
	return nil

}

// Remove tailscale funnel when no active events remain
func RemoveFunnel() error {
	// Check if any other active events exist
	events, err := models.GetAllActiveEvents()
	if err != nil {
		log.Printf("Failed to check active events before funnel removal: %v", err)
		return err
	}

	if len(events) > 0 {
		log.Printf("Keeping funnel active: %d event(s) still active", len(events))
		return nil
	}

	// Safe to remove - no active events
	cmd := exec.Command(
		"docker", "exec", "ts",
		"tailscale", "funnel", "off",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to remove funnel: %v - Output: %s", err, string(output))
		return fmt.Errorf("funnel removal failed: %w", err)
	}

	log.Printf("All events inactive - funnel removed")
	return nil
}
