package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient wires a Client to talk to the given test server URL instead of
// the real F5 XC endpoint.
func newTestClient(serverURL, apiKey string) *Client {
	return &Client{
		tenant: "", // unused — URL is overridden in the test server
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL: serverURL,
	}
}

// TestFetchEvents_1hWindow checks that the 1h window computes a start_time
// approximately 1 hour before now.
func TestFetchEvents_1hWindow(t *testing.T) {
	before := time.Now().UTC()

	var capturedBody eventsRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(EventsResponse{Events: []SecurityEvent{}})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-key")
	_, err := client.FetchEvents(context.Background(), "test-ns", "my-lb", "1h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	startTime, err := time.Parse(time.RFC3339, capturedBody.StartTime)
	if err != nil {
		t.Fatalf("parse start_time: %v", err)
	}

	want := before.Add(-1 * time.Hour)
	diff := startTime.Sub(want)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("start_time %v not within 5s of expected %v (diff %v)", startTime, want, diff)
	}
}

// TestFetchEvents_24hWindow checks that the 24h window computes a start_time
// approximately 24 hours before now.
func TestFetchEvents_24hWindow(t *testing.T) {
	before := time.Now().UTC()

	var capturedBody eventsRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(EventsResponse{Events: []SecurityEvent{}})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "test-key")
	_, err := client.FetchEvents(context.Background(), "test-ns", "", "24h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	startTime, err := time.Parse(time.RFC3339, capturedBody.StartTime)
	if err != nil {
		t.Fatalf("parse start_time: %v", err)
	}

	want := before.Add(-24 * time.Hour)
	diff := startTime.Sub(want)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("start_time %v not within 5s of expected %v (diff %v)", startTime, want, diff)
	}
}

// TestFetchEvents_AuthHeader verifies the APIToken header is sent.
func TestFetchEvents_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&eventsRequest{})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(EventsResponse{})
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "my-secret-key")
	client.FetchEvents(context.Background(), "ns", "", "1h")

	if gotAuth != "APIToken my-secret-key" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "APIToken my-secret-key")
	}
}

// TestFetchEvents_Non200 checks that a non-200 response returns an error.
func TestFetchEvents_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL, "bad-key")
	_, err := client.FetchEvents(context.Background(), "ns", "", "1h")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error %q does not mention status 401", err.Error())
	}
}

// TestFetchEvents_InvalidWindow checks that an unknown window string returns an error.
func TestFetchEvents_InvalidWindow(t *testing.T) {
	client := &Client{httpClient: &http.Client{}}
	_, err := client.FetchEvents(context.Background(), "ns", "", "7d")
	if err == nil {
		t.Fatal("expected error for invalid window, got nil")
	}
}
