package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const tailscaleSocket = "/var/run/tailscale/tailscaled.sock"

type TailscaleClient struct {
	client *http.Client
}

func NewTailscaleClient() *TailscaleClient {
	return &TailscaleClient{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", tailscaleSocket)
				},
			},
			Timeout: 5 * time.Second,
		},
	}
}

type tailscaleStatus struct {
	Self struct {
		DNSName string `json:"DNSNAME"`
	} `json:"Self"`
}

func (tc *TailscaleClient) GetFunnelURL(slug string) (string, error) {
	// Query Tailscale's local API for status
	resp, err := tc.client.Get("http://local-tailscaled.sock/localapi/v0/status")
	if err != nil {
		return "", fmt.Errorf("failed to query tailscale API: %w", err)
	}
	defer resp.Body.Close()

	// If http returns bad status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tailscale API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response: %w", err)
	}

	var status tailscaleStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return "", fmt.Errorf("failed to parse status %w", err)
	}

	if status.Self.DNSName == "" {
		return "", fmt.Errorf("tailscale not connected: no DNS name available")
	}

	// Remove trailing dot from DNSName
	hostname := strings.TrimSuffix(status.Self.DNSName, ".")

	// Construct funnel URL
	funnelURL := fmt.Sprintf("https://%s/events/%s", hostname, slug)

	return funnelURL, nil
}

func (tc *TailscaleClient) CheckFunnelActive() (bool, error) {
	// Query serve config endpoint
	resp, err := tc.client.Get("http://local-tailscaled.sock/localapi/v0/serve-config")
	if err != nil {
		// If endpoint unavailable, funnel likely not active
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read serve config %w", err)
	}

	// Basic check if port 8080 is in the returned config
	return strings.Contains(string(body), "8080"), nil
}
