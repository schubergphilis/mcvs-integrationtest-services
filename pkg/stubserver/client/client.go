// client package provides a client for interacting with a stub server.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/schubergphilis/mcvs-integrationtest-services/internal/app/stubserver"
	"github.com/schubergphilis/mcvs-integrationtest-services/pkg/stubserver/models"
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

func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

func closeResponseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		err := resp.Body.Close()
		if err != nil {
			log.Println("failed to close response body")
		}
	}
}

// HealthCheck checks if the stub server is healthy.
func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s%s", c.baseURL, stubserver.HealthEndpoint)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to perform health check: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	return nil
}

// AddResponse adds a new response to the stub server.
func (c *Client) AddResponse(ctx context.Context, request models.EndpointRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s%s", c.baseURL, stubserver.BaseURLPath, stubserver.ResponsesEndpoint)
	headers := map[string]string{"Content-Type": "application/json"}

	resp, err := c.doRequest(ctx, http.MethodPost, url, bytes.NewBuffer(data), headers)
	if err != nil {
		return fmt.Errorf("failed to add response: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		var errorResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return fmt.Errorf("failed with status code %d", resp.StatusCode)
		}

		return fmt.Errorf("failed to add response: %s", errorResp.Error)
	}

	return nil
}

// GetAllResponses retrieves all responses from the stub server.
func (c *Client) GetAllResponses(ctx context.Context) ([]models.EndpointResponse, error) {
	url := fmt.Sprintf("%s%s%s", c.baseURL, stubserver.BaseURLPath, stubserver.ResponsesEndpoint)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed with status code %d", resp.StatusCode)
	}

	var listResponse models.EndpointListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return listResponse.Endpoints, nil
}

// DeleteAllResponses deletes all responses from the stub server.
func (c *Client) DeleteAllResponses(ctx context.Context) error {
	url := fmt.Sprintf("%s%s%s", c.baseURL, stubserver.BaseURLPath, stubserver.ResponsesEndpoint)

	resp, err := c.doRequest(ctx, http.MethodDelete, url, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete responses: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed with status code %d", resp.StatusCode)
	}

	return nil
}

// SendRequest sends a request to a configured endpoint.
func (c *Client) SendRequest(ctx context.Context, method, path string, queryParams, headers map[string]string, body io.Reader) (*http.Response, error) {
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

	resp, err := c.doRequest(ctx, method, parsedURL.String(), body, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}
