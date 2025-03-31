package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	// given
	handler := newHandler()

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodGet, healthEndpoint, nil)

	// then
	handler.health(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
}

func TestHandleResponsesDelete(t *testing.T) {
	// given
	handler := newHandler()
	handler.endpoints["/test"] = "test"
	assert.Len(t, handler.endpoints, 1)

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodDelete, baseURLPath+responsesEndpoint, nil)

	// then
	handler.handleResponses(httptestRecorder, httptestRequest)
	assert.Len(t, handler.endpoints, 0)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
}

func TestHandleResponsesPost(t *testing.T) {
	// given
	handler := newHandler()
	assert.Len(t, handler.endpoints, 0)

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodPost, baseURLPath+responsesEndpoint, bytes.NewBufferString(`{"path": "/test", "response": {"foo": "bar"}}`))
	httptestRequest.Header.Set("Content-Type", "application/json")

	// then
	handler.handleResponses(httptestRecorder, httptestRequest)
	assert.Len(t, handler.endpoints, 1)
	b, err := json.Marshal(handler.endpoints["/test"])
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"foo":"bar"}`), b)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
}

func TestHandleResponsesInvalidMethod(t *testing.T) {
	// given
	handler := newHandler()

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodPatch, baseURLPath+responsesEndpoint, nil)

	// then
	handler.handleResponses(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusMethodNotAllowed, httptestRecorder.Code)
}

func TestCatchAllHandler(t *testing.T) {
	// given
	handler := newHandler()

	response := struct {
		Foo string `json:"foo"`
	}{
		Foo: "bar",
	}

	handler.endpoints["/test"] = response

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodGet, "/test", nil)

	// then
	handler.catchAll(httptestRecorder, httptestRequest)
	assert.Equal(t, []byte(`{"foo":"bar"}`), httptestRecorder.Body.Bytes())
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
}

func TestCatchAllHandlerNotFound(t *testing.T) {
	// given
	handler := newHandler()

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodGet, "/test", nil)

	// then
	handler.catchAll(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusNotFound, httptestRecorder.Code)
}

func TestListHandler(t *testing.T) {
	// given
	handler := newHandler()
	handler.endpoints["/foo"] = "bar"
	handler.endpoints["/bar"] = "foo"

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodGet, baseURLPath+responsesEndpoint, nil)

	// then
	handler.handleResponses(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
	assert.Len(t, handler.endpoints, 2)
	assert.Equal(t, []byte(`{"/bar":"foo","/foo":"bar"}`), httptestRecorder.Body.Bytes())
}

func TestListHandlerInvalidMethod(t *testing.T) {
	// given
	handler := newHandler()
	handler.endpoints["/bad"] = func() {}

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodGet, baseURLPath+responsesEndpoint, nil)

	// then
	handler.handleResponses(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusInternalServerError, httptestRecorder.Code)
}

func TestListHandlerInvalidEndpointsMap(t *testing.T) {
	// given
	handler := newHandler()
	handler.endpoints["/bad"] = func() {}

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest(http.MethodGet, baseURLPath+responsesEndpoint, nil)

	// then
	handler.getAllResponses(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusInternalServerError, httptestRecorder.Code)
}

func TestLogRequestContext(t *testing.T) {
	helperLogRequestContextBasicGetRequest(t)

	basicGETRequest(t)

	requestWithHeaderIncludingAuthorization(t)

	requestWithQueryParameters(t)

	requestWithBody(t)

	requestWithLargeBodyThatExceedsLimit(t)
}

func requestWithLargeBodyThatExceedsLimit(t *testing.T) {
	t.Helper()
	t.Run("Request with large body", func(t *testing.T) {
		// Create a body that's larger than the 10KB limit
		largeBody := strings.Repeat("X", 15*1024) // 15KB
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/data", bytes.NewBufferString(largeBody))

		result := logRequestContext(req)

		// Verify body is truncated (should contain some of the X's but not all 15KB)
		assertContains(t, result, "Body Content:")
		assertContains(t, result, "X") // Should have some content

		// The body in the result should be shorter than the original large body
		bodyStartIndex := strings.Index(result, "Body Content:\n") + len("Body Content:\n")
		bodyContent := result[bodyStartIndex:]
		if len(bodyContent) >= len(largeBody) {
			t.Errorf("Body was not limited: found %d bytes", len(bodyContent))
		}
	})
}

func requestWithBody(t *testing.T) {
	t.Helper()
	t.Run("Request with body", func(t *testing.T) {
		body := `{"username":"testuser","password":"secret"}`
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/login", bytes.NewBufferString(body))

		result := logRequestContext(req)

		// Verify body is logged
		assertContains(t, result, "Body Content:")
		assertContains(t, result, body)
	})
}

func requestWithQueryParameters(t *testing.T) {
	t.Helper()
	t.Run("Request with query parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/test?name=John&age=30&filter=active", nil)

		result := logRequestContext(req)

		// Verify query parameters are logged
		assertContains(t, result, "Query Parameters:")
		assertContains(t, result, "name: John")
		assertContains(t, result, "age: 30")
		assertContains(t, result, "filter: active")
	})
}

func requestWithHeaderIncludingAuthorization(t *testing.T) {
	t.Helper()
	t.Run("Request with Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer token123456")
		req.Header.Set("X-Custom-Header", "custom-value")

		result := logRequestContext(req)

		// Verify regular headers are logged normally
		assertContains(t, result, "Content-Type: application/json")
		assertContains(t, result, "X-Custom-Header: custom-value")

		// Verify Authorization header is redacted
		assertContains(t, result, "Authorization: *****")
		assertNotContains(t, result, "Bearer token123456")
	})
}

func basicGETRequest(t *testing.T) {
	t.Helper()
	t.Run("logs basic request info", func(t *testing.T) {
		// Setup test request
		req := httptest.NewRequest(http.MethodGet, "http://example.com/test?param=value", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer token123")
		req.RemoteAddr = "192.168.1.100:12345"

		// Call the function
		result := logRequestContext(req)

		// Assert results
		if !strings.Contains(result, "Request Method: GET") {
			t.Error("Missing request method in log")
		}
		if !strings.Contains(result, "Absolute URL: http://example.com/test?param=value") {
			t.Error("Missing or incorrect URL in log")
		}
		if !strings.Contains(result, "Absolute Path: /test") {
			t.Error("Missing or incorrect path in log")
		}
		if !strings.Contains(result, "Host: example.com") {
			t.Error("Missing host in log")
		}
		if !strings.Contains(result, "Remote Address: 192.168.1.100:12345") {
			t.Error("Missing remote address in log")
		}
		if !strings.Contains(result, "Content-Type: application/json") {
			t.Error("Missing regular header in log")
		}
		if !strings.Contains(result, "Authorization: *****") {
			t.Error("Authorization header not properly masked")
		}
		if strings.Contains(result, "Bearer token123") {
			t.Error("Authorization token should not appear in log")
		}
		if !strings.Contains(result, "param: value") {
			t.Error("Missing query parameter in log")
		}
	})
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected string to contain '%s', but it didn't.\nGot: %s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("Expected string to NOT contain '%s', but it did.", needle)
	}
}

func TestCompleteServerFlow(t *testing.T) {
	// Start the test server
	server, h := startTestServer(t)
	defer server.Close()

	// Test 1: Check health endpoint using direct handler call
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, healthEndpoint, nil)
	h.health(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 2: Add a response using direct handler call
	payload := EndpointConfigurationRequest{
		Path:     "/api/products",
		Response: map[string]interface{}{"products": []string{"product1", "product2"}},
	}
	respBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to serialize payload: %v", err)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, baseURLPath+responsesEndpoint, bytes.NewBuffer(respBody))
	req.Header.Set("Content-Type", "application/json")
	h.handleResponses(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 3: Verify the response is available using direct handler call
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/products", nil)
	h.catchAll(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&responseBody)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	products, ok := responseBody["products"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(products))

	// Test 4: List all responses
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, baseURLPath+responsesEndpoint, nil)
	h.handleResponses(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 5: Delete all responses using direct handler call
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, baseURLPath+responsesEndpoint, nil)
	h.handleResponses(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 6: Verify the response was removed using direct handler call
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/products", nil)
	h.catchAll(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// startTestServer creates and runs a test server with all handlers
// returns an instance of the test server and a pointer to the handler for state verification.
func startTestServer(t *testing.T) (*httptest.Server, *handler) {
	t.Helper()
	h := newHandler()

	// Define multiplexer for all paths
	mux := http.NewServeMux()
	mux.HandleFunc(healthEndpoint, h.health)
	mux.HandleFunc(baseURLPath+responsesEndpoint, h.handleResponses)
	mux.HandleFunc("/", h.catchAll)

	// Create and start the test server
	server := httptest.NewServer(mux)

	// Add cleanup that will close the server after the test completes
	t.Cleanup(func() {
		server.Close()
	})

	return server, h
}

func helperLogRequestContextBasicGetRequest(t *testing.T) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	req.RemoteAddr = "192.168.1.100:12345"

	result := logRequestContext(req)

	testCases := []struct {
		substring    string
		errorMsg     string
		checkContain bool
	}{
		{"Request Method: GET", "Missing request method in log", true},
		{"Absolute URL: http://example.com/test?param=value", "Missing or incorrect URL in log", true},
		{"Absolute Path: /test", "Missing or incorrect path in log", true},
		{"Host: example.com", "Missing host in log", true},
		{"Remote Address: 192.168.1.100:12345", "Missing remote address in log", true},
		{"Content-Type: application/json", "Missing regular header in log", true},
		{"Authorization: *****", "Authorization header not properly masked", true},
		{"Bearer token123", "Authorization token should not appear in log", false},
		{"param: value", "Missing query parameter in log", true},
	}

	for _, tc := range testCases {
		if tc.checkContain && !strings.Contains(result, tc.substring) {
			t.Error(tc.errorMsg)
		}
		if !tc.checkContain && strings.Contains(result, tc.substring) {
			t.Error(tc.errorMsg)
		}
	}
}
