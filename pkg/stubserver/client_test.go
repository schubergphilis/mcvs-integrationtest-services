package stubserver_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/schubergphilis/mcvs-integrationtest-services/internal/stubserver"
	stub_server_client "github.com/schubergphilis/mcvs-integrationtest-services/pkg/stubserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StubServerTestSuite struct {
	suite.Suite
	server     *stubserver.Server
	testServer *httptest.Server
	client     *stub_server_client.Client
}

func (s *StubServerTestSuite) SetupTest() {
	gin.SetMode(gin.DebugMode)
	s.server = stubserver.NewServer()
	s.testServer = httptest.NewServer(s.server.Router)
	s.client = stub_server_client.NewClient(s.testServer.URL, nil)
}

func (s *StubServerTestSuite) TearDownTest() {
	s.testServer.Close()
}

func TestStubServerSuite(t *testing.T) {
	suite.Run(t, new(StubServerTestSuite))
}

func (s *StubServerTestSuite) TestHealthCheck() {
	err := s.client.HealthCheck(s.T().Context())
	assert.NoError(s.T(), err)
}

func (s *StubServerTestSuite) TestAddResponse() {
	testRequest := stub_server_client.EndpointRequest{
		Path:               "/test",
		HTTPMethod:         http.MethodGet,
		ResponseBody:       "test response",
		ResponseStatusCode: http.StatusOK,
		ResponseHeaders:    map[string]string{"Content-Type": "text/plain"},
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	responses, err := s.client.GetAllResponses(s.T().Context())
	assert.NoError(s.T(), err)
	assert.Len(s.T(), responses, 1)
	assert.Equal(s.T(), testRequest.Path, responses[0].Path)
	assert.Equal(s.T(), testRequest.HTTPMethod, responses[0].HTTPMethod)
	assert.Equal(s.T(), testRequest.ResponseBody, responses[0].ResponseBody)
}

func (s *StubServerTestSuite) TestAddResponseWithInvalidData() {
	testCases := []struct {
		name          string
		request       stub_server_client.EndpointRequest
		errorContains string
	}{
		{
			name: "missing path",
			request: stub_server_client.EndpointRequest{
				HTTPMethod:         http.MethodGet,
				ResponseBody:       "test response",
				ResponseStatusCode: http.StatusOK,
			},
			errorContains: "Path and HTTP method are required",
		},
		{
			name: "missing HTTP method",
			request: stub_server_client.EndpointRequest{
				Path:               "/test",
				ResponseBody:       "test response",
				ResponseStatusCode: http.StatusOK,
			},
			errorContains: "Path and HTTP method are required",
		},
		{
			name: "missing response body",
			request: stub_server_client.EndpointRequest{
				Path:               "/test",
				HTTPMethod:         http.MethodGet,
				ResponseStatusCode: http.StatusOK,
			},
			errorContains: "Response body is required",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.client.AddResponse(s.T().Context(), tc.request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorContains)
		})
	}
}

func (s *StubServerTestSuite) TestAddDuplicateEndpoint() {
	testRequest := stub_server_client.EndpointRequest{
		Path:               "/test",
		HTTPMethod:         http.MethodGet,
		ResponseBody:       "test response",
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	err = s.client.AddResponse(s.T().Context(), testRequest)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "endpoint already exists")
}

func (s *StubServerTestSuite) TestGetAllResponses() {
	responses := []stub_server_client.EndpointRequest{
		{
			Path:               "/test1",
			HTTPMethod:         http.MethodGet,
			ResponseBody:       "test response 1",
			ResponseStatusCode: http.StatusOK,
		},
		{
			Path:               "/test2",
			HTTPMethod:         http.MethodPost,
			ResponseBody:       "test response 2",
			ResponseStatusCode: http.StatusCreated,
		},
	}

	for _, resp := range responses {
		err := s.client.AddResponse(s.T().Context(), resp)
		assert.NoError(s.T(), err)
	}

	retrievedResponses, err := s.client.GetAllResponses(s.T().Context())
	assert.NoError(s.T(), err)
	assert.Len(s.T(), retrievedResponses, 2)

	paths := []string{retrievedResponses[0].Path, retrievedResponses[1].Path}
	assert.Contains(s.T(), paths, "/test1")
	assert.Contains(s.T(), paths, "/test2")
}

func (s *StubServerTestSuite) TestDeleteAllResponses() {
	testRequest := stub_server_client.EndpointRequest{
		Path:               "/test",
		HTTPMethod:         http.MethodGet,
		ResponseBody:       "test response",
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	responses, err := s.client.GetAllResponses(s.T().Context())
	assert.NoError(s.T(), err)
	assert.Len(s.T(), responses, 1)

	err = s.client.DeleteAllResponses(s.T().Context())
	assert.NoError(s.T(), err)

	responses, err = s.client.GetAllResponses(s.T().Context())
	assert.NoError(s.T(), err)
	assert.Len(s.T(), responses, 0)
}

func (s *StubServerTestSuite) TestSendRequestWithQueryParamsMatching() {
	testRequest := stub_server_client.EndpointRequest{
		Path:               "/api/v1/products",
		HTTPMethod:         http.MethodGet,
		QueryParamsToMatch: map[string]string{"page": "3", "limit": "25"},
		ResponseHeaders:    map[string]string{"Content-Type": "application/json"},
		ResponseBody:       `{"items":[],"page":3,"limit":25,"total":100}`,
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	resp, err := s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		"/api/v1/products",
		map[string]string{"page": "3", "limit": "25"},
		nil,
		nil,
	)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(s.T(), "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), `{"items":[],"page":3,"limit":25,"total":100}`, string(body))

	resp, err = s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		"/api/v1/products",
		map[string]string{"page": "2", "limit": "25"},
		nil,
		nil,
	)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StubServerTestSuite) TestSendRequestWithHeadersMatching() {
	testRequest := stub_server_client.EndpointRequest{
		Path:               "/api/v1/users",
		HTTPMethod:         http.MethodGet,
		HeadersToMatch:     map[string]string{"X-Page": "2", "X-Per-Page": "50"},
		ResponseHeaders:    map[string]string{"Content-Type": "application/json"},
		ResponseBody:       `{"users":[],"page":2,"per_page":50,"total":150}`,
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	resp, err := s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		"/api/v1/users",
		nil,
		map[string]string{"X-Page": "2", "X-Per-Page": "50"},
		nil,
	)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), `{"users":[],"page":2,"per_page":50,"total":150}`, string(body))

	resp, err = s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		"/api/v1/users",
		nil,
		map[string]string{"X-Page": "3", "X-Per-Page": "50"},
		nil,
	)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StubServerTestSuite) TestSendRequestWithBothHeadersAndQueryParams() {
	testRequest := stub_server_client.EndpointRequest{
		Path:               "/api/v1/orders",
		HTTPMethod:         http.MethodGet,
		QueryParamsToMatch: map[string]string{"sort": "created_at", "order": "desc"},
		HeadersToMatch:     map[string]string{"X-Page": "1", "X-Per-Page": "100"},
		ResponseHeaders:    map[string]string{"Content-Type": "application/json"},
		ResponseBody:       `{"orders":[],"page":1,"per_page":100,"sort":"created_at","order":"desc"}`,
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	resp, err := s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		"/api/v1/orders",
		map[string]string{"sort": "created_at", "order": "desc"},
		map[string]string{"X-Page": "1", "X-Per-Page": "100"},
		nil,
	)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), `{"orders":[],"page":1,"per_page":100,"sort":"created_at","order":"desc"}`, string(body))

	resp, err = s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		"/api/v1/orders",
		map[string]string{"sort": "created_at", "order": "asc"},
		map[string]string{"X-Page": "1", "X-Per-Page": "100"},
		nil,
	)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StubServerTestSuite) TestSendRequestWithBody() {
	createUserRequest := stub_server_client.EndpointRequest{
		Path:               "/api/users",
		HTTPMethod:         http.MethodPost,
		ResponseHeaders:    map[string]string{"Content-Type": "application/json"},
		ResponseBody:       `{"id": "456", "name": "New User", "status": "created"}`,
		ResponseStatusCode: http.StatusCreated,
	}

	err := s.client.AddResponse(s.T().Context(), createUserRequest)
	assert.NoError(s.T(), err)

	requestBody := map[string]string{
		"name":  "New User",
		"email": "user@example.com",
	}
	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(s.T(), err)

	resp, err := s.client.SendRequest(s.T().Context(),
		http.MethodPost,
		"/api/users",
		nil,
		map[string]string{"Content-Type": "application/json"},
		strings.NewReader(string(jsonBody)),
	)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	var responseData map[string]string
	err = json.Unmarshal(body, &responseData)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "456", responseData["id"])
	assert.Equal(s.T(), "New User", responseData["name"])
	assert.Equal(s.T(), "created", responseData["status"])
}

func (s *StubServerTestSuite) TestClientWithInvalidBaseURL() {
	client := stub_server_client.NewClient("http://localhost:99999", nil)

	err := client.HealthCheck(s.T().Context())
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err, "failed to perform health check")

	_, err = client.GetAllResponses(s.T().Context())
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "failed to get responses")

	err = client.DeleteAllResponses(s.T().Context())
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err, "failed to delete responses")

	err = client.AddResponse(s.T().Context(), stub_server_client.EndpointRequest{
		Path:               "/test",
		HTTPMethod:         http.MethodGet,
		ResponseBody:       "test",
		ResponseStatusCode: http.StatusOK,
	})
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err, "failed to add response")
}
