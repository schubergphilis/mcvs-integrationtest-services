package models

// EndpointRequest represents the request body for adding a new endpoint.
type EndpointRequest struct {
	Path               string            `json:"path"`
	HTTPMethod         string            `json:"httpMethod"`
	QueryParamsToMatch map[string]string `json:"queryParamsToMatch,omitempty"`
	HeadersToMatch     map[string]string `json:"headersToMatch,omitempty"`
	ResponseHeaders    map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody       string            `json:"responseBody"`
	ResponseStatusCode int               `json:"responseStatusCode"`
}

// EndpointListResponse EndpointListRequest represents the request body for listing endpoints.
type EndpointListResponse struct {
	Endpoints []EndpointResponse `json:"endpoints"`
}

// EndpointResponse represents the response body for an endpoint.
type EndpointResponse struct {
	Path               string            `json:"path"`
	HTTPMethod         string            `json:"httpMethod"`
	QueryParamsToMatch map[string]string `json:"queryParamsToMatch,omitempty"`
	HeadersToMatch     map[string]string `json:"headersToMatch,omitempty"`
	ResponseHeaders    map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody       string            `json:"responseBody"`
	ResponseStatusCode int               `json:"responseStatusCode"`
}

// ErrorResponse represents the error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}
