package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
