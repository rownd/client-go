package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient wraps the standard http.Client with additional functionality
type HTTPClient struct {
	client     *http.Client
	maxRetries int
}

// RequestOptions contains options for making HTTP requests
type RequestOptions struct {
	Headers map[string]string
	Query   map[string]string
	Timeout time.Duration
}

// NewHTTPClient creates a new HTTP client with the specified configuration
func NewHTTPClient(timeout time.Duration, maxRetries int) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
	}
}

// DoRequest performs an HTTP request with retries and error handling
func (c *HTTPClient) DoRequest(ctx context.Context, method, url string, body interface{}, opts *RequestOptions) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply options
	if opts != nil {
		// Add headers
		for key, value := range opts.Headers {
			req.Header.Set(key, value)
		}

		// Add query parameters
		q := req.URL.Query()
		for key, value := range opts.Query {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Set default headers
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Perform request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, lastErr = c.client.Do(req)
		if lastErr == nil {
			break
		}

		// Check if context is cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Wait before retrying
		if attempt < c.maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
	}

	return resp, nil
}

// DecodeResponse decodes the response body into the provided interface
func DecodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
