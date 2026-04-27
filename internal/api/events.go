package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// FetchEvents retrieves WAF/security events from the F5 XC API.
//
//   - namespace: F5 XC namespace (e.g. "s-iannetta")
//   - lbName:    virtual_host / HTTP LB name to filter on (empty = no filter)
//   - hours:     look-back window in hours (1–24)
func (c *Client) FetchEvents(ctx context.Context, namespace, lbName string, hours int) ([]SecurityEvent, error) {
	if hours < 1 || hours > 24 {
		return nil, fmt.Errorf("invalid hours %d: must be 1–24", hours)
	}
	now := time.Now().UTC()
	startTime := now.Add(-time.Duration(hours) * time.Hour)

	query := ""
	if lbName != "" {
		query = fmt.Sprintf(`{vh_name="%s"}`, lbName)
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

	fmt.Fprintf(os.Stderr, "[DEBUG] request body: %s\n", bodyBytes)

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

	fmt.Fprintf(os.Stderr, "[DEBUG] response status: %d\n[DEBUG] response body: %s\n", resp.StatusCode, respBytes)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("F5 XC API returned HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var eventsResp EventsResponse
	if err := json.Unmarshal(respBytes, &eventsResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	events := make([]SecurityEvent, 0, len(eventsResp.RawEvents))
	for i, raw := range eventsResp.RawEvents {
		var e SecurityEvent
		if err := json.Unmarshal([]byte(raw), &e); err != nil {
			return nil, fmt.Errorf("unmarshal event %d: %w", i, err)
		}
		events = append(events, e)
	}
	return events, nil
}
