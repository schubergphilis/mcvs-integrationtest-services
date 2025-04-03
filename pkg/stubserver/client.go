package stubserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

// Client represents an HTTP client for the stub server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new instance of the client with the given base URL.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// HealthCheck checks if the stub server is healthy.
func (c *Client) HealthCheck() error {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s%s", c.baseURL, HealthEndpoint))
	if err != nil {
		return fmt.Errorf("failed to perform health check: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println("failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	return nil
}

// AddResponse adds a new response to the stub server.
func (c *Client) AddResponse(request EndpointRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s%s%s", c.baseURL, BaseURLPath, ResponsesEndpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return fmt.Errorf("failed to add response: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println("failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return fmt.Errorf("failed with status code %d", resp.StatusCode)
		}

		return fmt.Errorf("failed to add response: %s", errorResp.Error)
	}

	return nil
}

// GetAllResponses retrieves all responses from the stub server.
func (c *Client) GetAllResponses() ([]EndpointResponse, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s%s%s", c.baseURL, BaseURLPath, ResponsesEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println("failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed with status code %d", resp.StatusCode)
	}

	var listResponse EndpointListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return listResponse.Endpoints, nil
}

// DeleteAllResponses deletes all responses from the stub server.
func (c *Client) DeleteAllResponses() error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s%s", c.baseURL, BaseURLPath, ResponsesEndpoint), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete responses: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println("failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed with status code %d", resp.StatusCode)
	}

	return nil
}

// SendRequest sends a request to a configured endpoint.
func (c *Client) SendRequest(method, path string, queryParams, headers map[string]string, body io.Reader) (*http.Response, error) {
	urlStr := fmt.Sprintf("%s%s", c.baseURL, path)
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if len(queryParams) > 0 {
		q := parsedURL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		parsedURL.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(method, parsedURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}
