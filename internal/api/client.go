package api

import (
	"net/http"
	"time"
)

// Client is an authenticated HTTP client for the F5 XC app_security events API.
type Client struct {
	tenant     string
	apiKey     string
	httpClient *http.Client
	// baseURL overrides the default F5 XC endpoint; used in tests.
	baseURL string
}

// NewClient creates a Client for the given tenant and API key.
// The underlying HTTP client has a 30-second timeout by default.
func NewClient(tenant, apiKey string) *Client {
	return &Client{
		tenant: tenant,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithTimeout returns a copy of the client with the HTTP timeout overridden.
func (c *Client) WithTimeout(d time.Duration) *Client {
	return &Client{
		tenant:  c.tenant,
		apiKey:  c.apiKey,
		baseURL: c.baseURL,
		httpClient: &http.Client{
			Timeout: d,
		},
	}
}

// WithBaseURL returns a copy of the client with the base URL overridden.
// Intended for testing — points requests at an httptest.Server instead of F5 XC.
func (c *Client) WithBaseURL(url string) *Client {
	return &Client{
		tenant:     c.tenant,
		apiKey:     c.apiKey,
		httpClient: c.httpClient,
		baseURL:    url,
	}
}

// WithAPIKey returns a copy of the client with the API key replaced.
// Used by web handlers when the key is supplied per-request via the UI.
func (c *Client) WithAPIKey(key string) *Client {
	return &Client{
		tenant:     c.tenant,
		apiKey:     key,
		httpClient: c.httpClient,
		baseURL:    c.baseURL,
	}
}

// WithTenant returns a copy of the client with the tenant replaced.
// Used by web handlers when the tenant is supplied per-request via the UI.
func (c *Client) WithTenant(tenant string) *Client {
	return &Client{
		tenant:     tenant,
		apiKey:     c.apiKey,
		httpClient: c.httpClient,
		baseURL:    c.baseURL,
	}
}
