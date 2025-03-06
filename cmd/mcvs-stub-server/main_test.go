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
	httptestRequest := httptest.NewRequest("GET", "/health", nil)

	// then
	handler.health(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
}

func TestResetHandler(t *testing.T) {
	// given
	handler := newHandler()
	handler.endpoints["/test"] = "test"
	assert.Len(t, handler.endpoints, 1)

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest("GET", "/reset", nil)

	// then
	handler.reset(httptestRecorder, httptestRequest)
	assert.Len(t, handler.endpoints, 0)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
}

func TestConfigureHandler(t *testing.T) {
	// given
	handler := newHandler()
	assert.Len(t, handler.endpoints, 0)

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest("POST", "/configure", bytes.NewBuffer([]byte(`{"path": "/test", "response": {"foo": "bar"}}`)))

	// then
	handler.configure(httptestRecorder, httptestRequest)
	assert.Len(t, handler.endpoints, 1)
	b, err := json.Marshal(handler.endpoints["/test"])
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"foo":"bar"}`), b)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
}

func TestConfigureHandlerInvalidMethod(t *testing.T) {
	// given
	handler := newHandler()

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest("GET", "/configure", nil)

	// then
	handler.configure(httptestRecorder, httptestRequest)
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
	httptestRequest := httptest.NewRequest("GET", "/test", nil)

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
	httptestRequest := httptest.NewRequest("GET", "/test", nil)

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
	httptestRequest := httptest.NewRequest("GET", "/list", nil)

	// then
	handler.list(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusOK, httptestRecorder.Code)
	assert.Len(t, handler.endpoints, 2)
	assert.Equal(t, []byte(`{"/bar":"foo","/foo":"bar"}`), httptestRecorder.Body.Bytes())
}

func TestListHandlerInvalidMethod(t *testing.T) {
	// given
	handler := newHandler()

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest("POST", "/list", nil)

	// then
	handler.list(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusMethodNotAllowed, httptestRecorder.Code)
}

func TestListHandlerInvalidEndpointsMap(t *testing.T) {
	// given
	handler := newHandler()
	handler.endpoints["/bad"] = func() {}

	// when
	httptestRecorder := httptest.NewRecorder()
	httptestRequest := httptest.NewRequest("GET", "/list", nil)

	// then
	handler.list(httptestRecorder, httptestRequest)
	assert.Equal(t, http.StatusInternalServerError, httptestRecorder.Code)
}

func TestLogRequestContext(t *testing.T) {
	// Test 1: Basic GET request
	t.Run("logs basic request info", func(t *testing.T) {
		// Setup test request
		req := httptest.NewRequest("GET", "http://example.com/test?param=value", nil)
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

	// Test 2: Request with headers including Authorization
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

	// Test 3: Request with query parameters
	t.Run("Request with query parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/test?name=John&age=30&filter=active", nil)

		result := logRequestContext(req)

		// Verify query parameters are logged
		assertContains(t, result, "Query Parameters:")
		assertContains(t, result, "name: John")
		assertContains(t, result, "age: 30")
		assertContains(t, result, "filter: active")
	})

	// Test 4: Request with body
	t.Run("Request with body", func(t *testing.T) {
		body := `{"username":"testuser","password":"secret"}`
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/login", bytes.NewBufferString(body))

		result := logRequestContext(req)

		// Verify body is logged
		assertContains(t, result, "Body Content:")
		assertContains(t, result, body)
	})

	// Test 5: Request with large body that exceeds limit
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
