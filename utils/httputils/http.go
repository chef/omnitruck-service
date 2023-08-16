package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// HTTPClient interface represents an HTTP client.
type HTTPClient interface {
	MakeRequest(requestMethod, url string, body interface{}) (*http.Response, []byte, error)
}

// Client is a wrapper around http.Client that adds some convenience methods for making requests and handling responses.
type Client struct {
	Client *http.Client
	Logger *log.Entry
}

// MockHTTPClient is a mock implementation of HTTPClient.
type MockHTTPClient struct {
	MakeRequestFunc func(requestMethod, url string, body interface{}) (*http.Response, []byte, error)
}

func (m *MockHTTPClient) MakeRequest(requestMethod, url string, body interface{}) (*http.Response, []byte, error) {
	return m.MakeRequestFunc(requestMethod, url, body)
}

// NewClient returns a new Client with sane defaults.
func NewClient(logger *log.Entry) *Client {
	return &Client{
		Client: http.DefaultClient,
		Logger: logger,
	}
}

// NewClientWithTimeout returns a new Client with the specified timeout.
func NewClientWithTimeout(timeout time.Duration, logger *log.Entry) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: timeout,
		},
		Logger: logger,
	}
}

// MakeRequest sends an HTTP request with the specified request method, URL, and body.
func (c *Client) MakeRequest(requestMethod, url string, body interface{}) (*http.Response, []byte, error) {
	var reader io.Reader

	// Marshal the request body to JSON if it's not nil
	if body != nil {
		requestBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		c.Logger.Debugf("Request body for MakeRequest: \n%s\n", string(requestBody))
		reader = bytes.NewReader(requestBody)
	}

	req, err := http.NewRequest(requestMethod, url, reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.Logger.Debugf("Response body for MakeRequest: \n%s\n", string(responseBody))
	return resp, responseBody, nil
}
