package stubserver

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResponseManager(t *testing.T) {
	rm := NewResponseManager()
	assert.NotNil(t, rm)
	assert.NotNil(t, rm.endpoints)
	assert.Empty(t, rm.endpoints)
}

func TestGetID(t *testing.T) {
	tests := []struct {
		name     string
		endpoint EndpointID
		expected string
	}{
		{
			name: "empty headers and params",
			endpoint: EndpointID{
				Path:               "/api/v1/test",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{},
			},
			expected: strings.ToLower("/api/v1/test:GET"),
		},
		{
			name: "pagination in headers",
			endpoint: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{"X-Page": "2", "X-Per-Page": "50"},
			},
			expected: strings.ToLower("/api/v1/users:GET:X-Page=2:X-Per-Page=50"),
		},
		{
			name: "pagination in query params",
			endpoint: EndpointID{
				Path:               "/api/v1/products",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "3", "limit": "25"},
				HeadersToMatch:     map[string]string{},
			},
			expected: strings.ToLower("/api/v1/products:GET:limit=25:page=3"),
		},
		{
			name: "sorting in query params with pagination in headers",
			endpoint: EndpointID{
				Path:               "/api/v1/orders",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"sort": "created_at", "order": "desc"},
				HeadersToMatch:     map[string]string{"X-Page": "1", "X-Per-Page": "100"},
			},
			expected: strings.ToLower("/api/v1/orders:GET:X-Page=1:X-Per-Page=100:order=desc:sort=created_at"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetID(&tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:funlen
func TestValidateEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		config        EndpointConfiguration
		expectedError bool
		errorMessage  string
	}{
		{
			name: "valid configuration",
			config: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:       "/api/v1/test",
					HTTPMethod: "GET",
				},
				ResponseBody:       "{\"status\":\"ok\"}",
				ResponseStatusCode: 200,
			},
			expectedError: false,
		},
		{
			name: "missing path",
			config: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:       "",
					HTTPMethod: "GET",
				},
				ResponseBody:       "{\"status\":\"ok\"}",
				ResponseStatusCode: 200,
			},
			expectedError: true,
			errorMessage:  "path and method are required",
		},
		{
			name: "missing method",
			config: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:       "/api/v1/test",
					HTTPMethod: "",
				},
				ResponseBody:       "{\"status\":\"ok\"}",
				ResponseStatusCode: 200,
			},
			expectedError: true,
			errorMessage:  "path and method are required",
		},
		{
			name: "missing response body",
			config: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:       "/api/v1/test",
					HTTPMethod: "GET",
				},
				ResponseBody:       "",
				ResponseStatusCode: 200,
			},
			expectedError: true,
			errorMessage:  "response body is required",
		},
		{
			name: "invalid HTTP method",
			config: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:       "/api/v1/test",
					HTTPMethod: "INVALID",
				},
				ResponseBody:       "{\"status\":\"ok\"}",
				ResponseStatusCode: 200,
			},
			expectedError: true,
			errorMessage:  "invalid HTTP method: INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpoint(tt.config)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddEndpoint(t *testing.T) {
	rm := NewResponseManager()

	// Test adding a valid endpoint
	validEndpoint := EndpointConfiguration{
		EndpointID: EndpointID{
			Path:               "/api/v1/test",
			HTTPMethod:         "GET",
			QueryParamsToMatch: map[string]string{},
			HeadersToMatch:     map[string]string{},
		},
		ResponseBody:       "{\"status\":\"ok\"}",
		ResponseStatusCode: 200,
	}
	err := rm.AddEndpoint(validEndpoint)
	assert.NoError(t, err)
	assert.Len(t, rm.endpoints, 1)

	// Test adding the same endpoint again (should fail)
	err = rm.AddEndpoint(validEndpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint already exists")
	assert.Len(t, rm.endpoints, 1)

	// Test adding an invalid endpoint
	invalidEndpoint := EndpointConfiguration{
		EndpointID: EndpointID{
			Path:       "",
			HTTPMethod: "GET",
		},
		ResponseBody:       "{\"status\":\"ok\"}",
		ResponseStatusCode: 200,
	}
	err = rm.AddEndpoint(invalidEndpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path and method are required")
	assert.Len(t, rm.endpoints, 1)
}

func TestGetEndpointByEndpointId(t *testing.T) {
	rm := NewResponseManager()

	// Add a test endpoint
	testEndpoint := EndpointConfiguration{
		EndpointID: EndpointID{
			Path:               "/api/v1/test",
			HTTPMethod:         "GET",
			QueryParamsToMatch: map[string]string{"param": "value"},
			HeadersToMatch:     map[string]string{"Header": "Value"},
		},
		ResponseBody:       "{\"status\":\"ok\"}",
		ResponseStatusCode: 200,
		ResponseHeaders:    map[string]string{"Content-Type": "application/json"},
	}
	err := rm.AddEndpoint(testEndpoint)
	assert.NoError(t, err)

	// Test retrieving the endpoint
	result, err := rm.GetEndpointByEndpointID(&testEndpoint.EndpointID)
	assert.NoError(t, err)
	assert.Equal(t, testEndpoint, result)

	// Test retrieving a non-existent endpoint
	nonExistentEndpoint := EndpointID{
		Path:       "/non/existent",
		HTTPMethod: "GET",
	}
	_, err = rm.GetEndpointByEndpointID(&nonExistentEndpoint)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint not found")
}

func TestGetAllEndpointConfigurations(t *testing.T) {
	rm := NewResponseManager()

	results := rm.GetAllEndpointConfigurations()
	assert.Empty(t, results)

	endpoints := []EndpointConfiguration{
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test1",
				HTTPMethod: "GET",
			},
			ResponseBody:       "{\"id\":1}",
			ResponseStatusCode: 200,
		},
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test2",
				HTTPMethod: "POST",
			},
			ResponseBody:       "{\"id\":2}",
			ResponseStatusCode: 201,
		},
	}

	for _, endpoint := range endpoints {
		err := rm.AddEndpoint(endpoint)
		assert.NoError(t, err)
	}

	results = rm.GetAllEndpointConfigurations()
	assert.Len(t, results, 2)

	found := make(map[string]bool)

	for _, result := range results {
		id := GetID(&result.EndpointID)
		found[id] = true
	}

	for _, endpoint := range endpoints {
		id := GetID(&endpoint.EndpointID)
		assert.True(t, found[id])
	}
}

func TestDeleteEndpointByEndpointId(t *testing.T) {
	rm := NewResponseManager()

	testEndpoint := EndpointConfiguration{
		EndpointID: EndpointID{
			Path:       "/api/v1/test",
			HTTPMethod: "GET",
		},
		ResponseBody:       "{\"status\":\"ok\"}",
		ResponseStatusCode: 200,
	}
	err := rm.AddEndpoint(testEndpoint)
	assert.NoError(t, err)
	assert.Len(t, rm.endpoints, 1)

	err = rm.DeleteEndpointByEndpointID(&testEndpoint.EndpointID)
	assert.NoError(t, err)
	assert.Empty(t, rm.endpoints)

	err = rm.DeleteEndpointByEndpointID(&testEndpoint.EndpointID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint not found")
}

func TestDeleteEndpointByPath(t *testing.T) {
	rm := NewResponseManager()

	endpoints := []EndpointConfiguration{
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test",
				HTTPMethod: "GET",
			},
			ResponseBody:       "{\"method\":\"GET\"}",
			ResponseStatusCode: 200,
		},
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test",
				HTTPMethod: "POST",
			},
			ResponseBody:       "{\"method\":\"POST\"}",
			ResponseStatusCode: 201,
		},
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/other",
				HTTPMethod: "GET",
			},
			ResponseBody:       "{\"path\":\"other\"}",
			ResponseStatusCode: 200,
		},
	}

	for _, endpoint := range endpoints {
		err := rm.AddEndpoint(endpoint)
		assert.NoError(t, err)
	}

	assert.Len(t, rm.endpoints, 3)

	err := rm.DeleteEndpointByPath("/api/v1/test")
	assert.NoError(t, err)

	assert.Len(t, rm.endpoints, 1)

	configs := rm.GetAllEndpointConfigurations()
	assert.Equal(t, "/api/v1/other", configs[0].EndpointID.Path)
}

func TestDeleteEndpointByPathAndMethod(t *testing.T) {
	rm := NewResponseManager()

	endpoints := []EndpointConfiguration{
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test",
				HTTPMethod: "GET",
			},
			ResponseBody:       "{\"method\":\"GET\"}",
			ResponseStatusCode: 200,
		},
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test",
				HTTPMethod: "POST",
			},
			ResponseBody:       "{\"method\":\"POST\"}",
			ResponseStatusCode: 201,
		},
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/other",
				HTTPMethod: "GET",
			},
			ResponseBody:       "{\"path\":\"other\"}",
			ResponseStatusCode: 200,
		},
	}

	for _, endpoint := range endpoints {
		err := rm.AddEndpoint(endpoint)
		assert.NoError(t, err)
	}

	assert.Len(t, rm.endpoints, 3)

	err := rm.DeleteEndpointByPathAndMethod("/api/v1/test", "GET")
	assert.NoError(t, err)

	assert.Len(t, rm.endpoints, 2)

	configs := rm.GetAllEndpointConfigurations()
	foundPaths := make(map[string]bool)
	foundMethods := make(map[string]bool)

	for _, config := range configs {
		foundPaths[config.EndpointID.Path] = true
		foundMethods[config.EndpointID.HTTPMethod] = true
	}

	assert.True(t, foundPaths["/api/v1/test"])
	assert.True(t, foundPaths["/api/v1/other"])
	assert.True(t, foundMethods["POST"])
	assert.True(t, foundMethods["GET"])

	for _, config := range configs {
		if config.EndpointID.Path == "/api/v1/test" {
			assert.NotEqual(t, "GET", config.EndpointID.HTTPMethod)
		}
	}
}

func TestDeleteAllEndpoints(t *testing.T) {
	rm := NewResponseManager()

	endpoints := []EndpointConfiguration{
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test1",
				HTTPMethod: "GET",
			},
			ResponseBody:       "{\"id\":1}",
			ResponseStatusCode: 200,
		},
		{
			EndpointID: EndpointID{
				Path:       "/api/v1/test2",
				HTTPMethod: "POST",
			},
			ResponseBody:       "{\"id\":2}",
			ResponseStatusCode: 201,
		},
	}

	for _, endpoint := range endpoints {
		err := rm.AddEndpoint(endpoint)
		assert.NoError(t, err)
	}

	assert.Len(t, rm.endpoints, 2)

	rm.DeleteAllEndpoints()
	assert.Empty(t, rm.endpoints)
}

//nolint:funlen
func TestMatchEndpoint(t *testing.T) {
	rm := NewResponseManager()

	endpoints := []EndpointConfiguration{
		{
			EndpointID: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
			},
			ResponseBody:       "{\"users\":[{\"id\":1}]}",
			ResponseStatusCode: 200,
		},
		{
			EndpointID: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "2", "limit": "20"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key2"},
			},
			ResponseBody:       "{\"users\":[{\"id\":2}]}",
			ResponseStatusCode: 200,
		},
		{
			EndpointID: EndpointID{
				Path:               "/api/v1/products",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"category": "electronics"},
				HeadersToMatch:     map[string]string{},
			},
			ResponseBody:       "{\"products\":[{\"id\":1}]}",
			ResponseStatusCode: 200,
		},
	}

	for _, endpoint := range endpoints {
		err := rm.AddEndpoint(endpoint)
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		request       EndpointID
		expectError   bool
		errorContains string
		expectedPath  string
	}{
		{
			name: "exact match for first endpoint",
			request: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
			},
			expectError:  false,
			expectedPath: "/api/v1/users",
		},
		{
			name: "exact match for second endpoint",
			request: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "2", "limit": "20"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key2"},
			},
			expectError:  false,
			expectedPath: "/api/v1/users",
		},
		{
			name: "partial match - more headers in request than endpoint",
			request: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1", "X-Extra": "value"},
			},
			expectError:  false,
			expectedPath: "/api/v1/users",
		},
		{
			name: "partial match - more query params in request than endpoint",
			request: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "10", "sort": "asc"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
			},
			expectError:  false,
			expectedPath: "/api/v1/users",
		},
		{
			name: "match with no query params or headers",
			request: EndpointID{
				Path:               "/api/v1/products",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{},
			},
			expectError:  false,
			expectedPath: "/api/v1/products",
		},
		{
			name: "match with some query params",
			request: EndpointID{
				Path:               "/api/v1/products",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"category": "electronics"},
				HeadersToMatch:     map[string]string{},
			},
			expectError:  false,
			expectedPath: "/api/v1/products",
		},
		{
			name: "no match - wrong path",
			request: EndpointID{
				Path:               "/api/v1/nonexistent",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{},
			},
			expectError:   true,
			errorContains: "no endpoints matched the given request",
		},
		{
			name: "no match - wrong method",
			request: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "POST",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{},
			},
			expectError:   true,
			errorContains: "no endpoints matched the given request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rm.MatchEndpoint(&tt.request)

			if tt.expectError {
				require.Error(t, err)

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}

				return
			}

			require.NoError(t, err)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedPath, result.EndpointID.Path)
		})
	}

	// Test matching logic with endpoints that would cause ambiguity
	// Add two endpoints with the same path/method but different query params/headers
	rm.DeleteAllEndpoints()

	ambiguousEndpoints := []EndpointConfiguration{
		{
			EndpointID: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
			},
			ResponseBody:       "{\"users\":[{\"id\":1}]}",
			ResponseStatusCode: 200,
		},
		{
			EndpointID: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
			},
			ResponseBody:       "{\"users\":[{\"id\":2}]}",
			ResponseStatusCode: 200,
		},
	}

	err := rm.AddEndpoint(ambiguousEndpoints[0])
	assert.NoError(t, err)

	err = rm.AddEndpoint(ambiguousEndpoints[1])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint already exists")
}

//nolint:funlen
func TestCalculateMatch(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       EndpointConfiguration
		requestToMatch EndpointID
		expectedScore  int
	}{
		{
			name: "exact match for all headers and query params",
			endpoint: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:               "/api/v1/users",
					HTTPMethod:         "GET",
					QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
					HeadersToMatch:     map[string]string{"X-API-Key": "key1", "Accept": "application/json"},
				},
			},
			requestToMatch: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1", "Accept": "application/json"},
			},
			expectedScore: 4, // 2 query params + 2 headers
		},
		{
			name: "partial match - only some headers match",
			endpoint: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:               "/api/v1/users",
					HTTPMethod:         "GET",
					QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
					HeadersToMatch:     map[string]string{"X-API-Key": "key1", "Accept": "application/json"},
				},
			},
			requestToMatch: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1", "Accept": "text/html"},
			},
			expectedScore: 3, // 2 query params + 1 header
		},
		{
			name: "partial match - only some query params match",
			endpoint: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:               "/api/v1/users",
					HTTPMethod:         "GET",
					QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
					HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
				},
			},
			requestToMatch: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "20"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
			},
			expectedScore: 2, // 1 query param + 1 header
		},
		{
			name: "additional params in request don't affect score",
			endpoint: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:               "/api/v1/users",
					HTTPMethod:         "GET",
					QueryParamsToMatch: map[string]string{"page": "1"},
					HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
				},
			},
			requestToMatch: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "1", "limit": "10", "sort": "desc"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key1", "Accept": "application/json"},
			},
			expectedScore: 2, // 1 query param + 1 header
		},
		{
			name: "no matches",
			endpoint: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:               "/api/v1/users",
					HTTPMethod:         "GET",
					QueryParamsToMatch: map[string]string{"page": "1", "limit": "10"},
					HeadersToMatch:     map[string]string{"X-API-Key": "key1"},
				},
			},
			requestToMatch: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "2", "limit": "20"},
				HeadersToMatch:     map[string]string{"X-API-Key": "key2"},
			},
			expectedScore: 0,
		},
		{
			name: "empty headers and query params",
			endpoint: EndpointConfiguration{
				EndpointID: EndpointID{
					Path:               "/api/v1/users",
					HTTPMethod:         "GET",
					QueryParamsToMatch: map[string]string{},
					HeadersToMatch:     map[string]string{},
				},
			},
			requestToMatch: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{},
			},
			expectedScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateMatch(&tt.endpoint, &tt.requestToMatch)
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}
