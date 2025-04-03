package stubserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			expected: "/api/v1/test:GET",
		},
		{
			name: "with headers",
			endpoint: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "POST",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{"Content-Type": "application/json"},
			},
			expected: "/api/v1/users:POST:Content-Type=application/json",
		},
		{
			name: "with query params",
			endpoint: EndpointID{
				Path:               "/api/v1/search",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"q": "test"},
				HeadersToMatch:     map[string]string{},
			},
			expected: "/api/v1/search:GET:q=test",
		},
		{
			name: "with both headers and params",
			endpoint: EndpointID{
				Path:               "/api/v1/products",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"category": "electronics", "sort": "price"},
				HeadersToMatch:     map[string]string{"Authorization": "Bearer token", "Accept": "application/json"},
			},
			expected: "/api/v1/products:GET:Authorization=Bearer token:Accept=application/json:category=electronics:sort=price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetID(&tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetId(t *testing.T) {
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
			expected: "/api/v1/test:GET",
		},
		{
			name: "pagination in headers",
			endpoint: EndpointID{
				Path:               "/api/v1/users",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{},
				HeadersToMatch:     map[string]string{"X-Page": "2", "X-Per-Page": "50"},
			},
			expected: "/api/v1/users:GET:X-Page=2:X-Per-Page=50",
		},
		{
			name: "pagination in query params",
			endpoint: EndpointID{
				Path:               "/api/v1/products",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"page": "3", "limit": "25"},
				HeadersToMatch:     map[string]string{},
			},
			expected: "/api/v1/products:GET:page=3:limit=25",
		},
		{
			name: "sorting in query params with pagination in headers",
			endpoint: EndpointID{
				Path:               "/api/v1/orders",
				HTTPMethod:         "GET",
				QueryParamsToMatch: map[string]string{"sort": "created_at", "order": "desc"},
				HeadersToMatch:     map[string]string{"X-Page": "1", "X-Per-Page": "100"},
			},
			expected: "/api/v1/orders:GET:X-Page=1:X-Per-Page=100:sort=created_at:order=desc",
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

	// Test with empty manager
	results := rm.GetAllEndpointConfigurations()
	assert.Empty(t, results)

	// Add multiple endpoints
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

	// Test getting all endpoints
	results = rm.GetAllEndpointConfigurations()
	assert.Len(t, results, 2)

	// Verify that all added endpoints are in the results
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

	// Add a test endpoint
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

	// Test deleting the endpoint
	err = rm.DeleteEndpointByEndpointID(&testEndpoint.EndpointID)
	assert.NoError(t, err)
	assert.Empty(t, rm.endpoints)

	// Test deleting a non-existent endpoint
	err = rm.DeleteEndpointByEndpointID(&testEndpoint.EndpointID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint not found")
}

func TestDeleteEndpointByPath(t *testing.T) {
	rm := NewResponseManager()

	// Add multiple endpoints with the same path but different methods
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

	// Test deleting endpoints by path
	err := rm.DeleteEndpointByPath("/api/v1/test")
	assert.NoError(t, err)

	// Should have only one endpoint left
	assert.Len(t, rm.endpoints, 1)

	// Verify the remaining endpoint
	configs := rm.GetAllEndpointConfigurations()
	assert.Equal(t, "/api/v1/other", configs[0].EndpointID.Path)
}

func TestDeleteEndpointByPathAndMethod(t *testing.T) {
	rm := NewResponseManager()

	// Add multiple endpoints
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

	// Test deleting endpoint by path and method
	err := rm.DeleteEndpointByPathAndMethod("/api/v1/test", "GET")
	assert.NoError(t, err)

	// Should have two endpoints left
	assert.Len(t, rm.endpoints, 2)

	// Verify the remaining endpoints
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

	// Verify no GET method for /api/v1/test path
	for _, config := range configs {
		if config.EndpointID.Path == "/api/v1/test" {
			assert.NotEqual(t, "GET", config.EndpointID.HTTPMethod)
		}
	}
}

func TestDeleteAllEndpoints(t *testing.T) {
	rm := NewResponseManager()

	// Add multiple endpoints
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

	// Test deleting all endpoints
	rm.DeleteAllEndpoints()
	assert.Empty(t, rm.endpoints)
}
