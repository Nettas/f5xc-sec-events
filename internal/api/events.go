package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchEvents retrieves WAF/security events from the F5 XC API.
//
//   - namespace: F5 XC namespace (e.g. "s-iannetta")
//   - lbName:    virtual_host / HTTP LB name to filter on (empty = no filter)
//   - window:    "1h" or "24h"
func (c *Client) FetchEvents(ctx context.Context, namespace, lbName, window string) ([]SecurityEvent, error) {
	now := time.Now().UTC()
	var startTime time.Time

	switch window {
	case "1h":
		startTime = now.Add(-1 * time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	default:
		return nil, fmt.Errorf("invalid window %q: must be \"1h\" or \"24h\"", window)
	}

	query := ""
	if lbName != "" {
		query = fmt.Sprintf("{virtual_host=%q}", lbName)
	}

	reqBody := eventsRequest{
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   now.Format(time.RFC3339),
		Namespace: namespace,
		Query:     query,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	var url string
	if c.baseURL != "" {
		url = c.baseURL
	} else {
		url = fmt.Sprintf(
			"https://%s.console.ves.volterra.io/api/data/namespaces/%s/app_security/events",
			c.tenant, namespace,
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "APIToken "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("F5 XC API returned HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var eventsResp EventsResponse
	if err := json.Unmarshal(respBytes, &eventsResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return eventsResp.Events, nil
}
