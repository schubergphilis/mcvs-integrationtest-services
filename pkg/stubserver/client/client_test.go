package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/schubergphilis/mcvs-integrationtest-services/internal/app/stubserver"
	"github.com/schubergphilis/mcvs-integrationtest-services/pkg/stubserver/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testPath         = "/test"
	testResponseBody = "test response"

	productsPath = "/api/v1/products"
	usersPath    = "/api/v1/users"
	ordersPath   = "/api/v1/orders"

	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
	pageHeader        = "X-Page"
	perPageHeader     = "X-Per-Page"

	pageParam      = "page"
	limitParam     = "limit"
	sortParam      = "sort"
	orderParam     = "order"
	createdAtValue = "created_at"
)

type StubServerTestSuite struct {
	suite.Suite

	server     *stubserver.Server
	testServer *httptest.Server
	client     *Client
}

func (s *StubServerTestSuite) SetupTest() {
	gin.SetMode(gin.DebugMode)

	s.server = stubserver.NewServer()
	s.testServer = httptest.NewServer(s.server.Router)
	s.client = NewClient(s.testServer.URL, nil)
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
	testRequest := models.EndpointRequest{
		Path:               testPath,
		HTTPMethod:         http.MethodGet,
		ResponseBody:       testResponseBody,
		ResponseStatusCode: http.StatusOK,
		ResponseHeaders:    map[string]string{contentTypeHeader: "text/plain"},
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
		request       models.EndpointRequest
		errorContains string
	}{
		{
			name: "missing path",
			request: models.EndpointRequest{
				HTTPMethod:         http.MethodGet,
				ResponseBody:       testResponseBody,
				ResponseStatusCode: http.StatusOK,
			},
			errorContains: "Path and HTTP method are required",
		},
		{
			name: "missing HTTP method",
			request: models.EndpointRequest{
				Path:               testPath,
				ResponseBody:       testResponseBody,
				ResponseStatusCode: http.StatusOK,
			},
			errorContains: "Path and HTTP method are required",
		},
		{
			name: "missing response body",
			request: models.EndpointRequest{
				Path:               testPath,
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
	testRequest := models.EndpointRequest{
		Path:               testPath,
		HTTPMethod:         http.MethodGet,
		ResponseBody:       testResponseBody,
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	err = s.client.AddResponse(s.T().Context(), testRequest)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "endpoint already exists")
}

func (s *StubServerTestSuite) TestGetAllResponses() {
	responses := []models.EndpointRequest{
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
	testRequest := models.EndpointRequest{
		Path:               testPath,
		HTTPMethod:         http.MethodGet,
		ResponseBody:       testResponseBody,
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
	testRequest := models.EndpointRequest{
		Path:               productsPath,
		HTTPMethod:         http.MethodGet,
		QueryParamsToMatch: map[string]string{pageParam: "3", limitParam: "25"},
		ResponseHeaders:    map[string]string{contentTypeHeader: jsonContentType},
		ResponseBody:       `{"items":[],"page":3,"limit":25,"total":100}`,
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	resp, err := s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		productsPath,
		map[string]string{pageParam: "3", limitParam: "25"},
		nil,
		nil,
	)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(s.T(), jsonContentType, resp.Header.Get(contentTypeHeader))

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), `{"items":[],"page":3,"limit":25,"total":100}`, string(body))

	resp, err = s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		productsPath,
		map[string]string{pageParam: "2", limitParam: "25"},
		nil,
		nil,
	)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StubServerTestSuite) TestSendRequestWithHeadersMatching() {
	testRequest := models.EndpointRequest{
		Path:               usersPath,
		HTTPMethod:         http.MethodGet,
		HeadersToMatch:     map[string]string{pageHeader: "2", perPageHeader: "50"},
		ResponseHeaders:    map[string]string{contentTypeHeader: jsonContentType},
		ResponseBody:       `{"users":[],"page":2,"per_page":50,"total":150}`,
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	resp, err := s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		usersPath,
		nil,
		map[string]string{pageHeader: "2", perPageHeader: "50"},
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
		usersPath,
		nil,
		map[string]string{pageHeader: "3", perPageHeader: "50"},
		nil,
	)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StubServerTestSuite) TestSendRequestWithBothHeadersAndQueryParams() {
	testRequest := models.EndpointRequest{
		Path:               ordersPath,
		HTTPMethod:         http.MethodGet,
		QueryParamsToMatch: map[string]string{sortParam: createdAtValue, orderParam: "desc"},
		HeadersToMatch:     map[string]string{pageHeader: "1", perPageHeader: "100"},
		ResponseHeaders:    map[string]string{contentTypeHeader: jsonContentType},
		ResponseBody:       `{"orders":[],"page":1,"per_page":100,"sort":"created_at","order":"desc"}`,
		ResponseStatusCode: http.StatusOK,
	}

	err := s.client.AddResponse(s.T().Context(), testRequest)
	assert.NoError(s.T(), err)

	resp, err := s.client.SendRequest(s.T().Context(),
		http.MethodGet,
		ordersPath,
		map[string]string{sortParam: createdAtValue, orderParam: "desc"},
		map[string]string{pageHeader: "1", perPageHeader: "100"},
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
		ordersPath,
		map[string]string{sortParam: createdAtValue, orderParam: "asc"},
		map[string]string{pageHeader: "1", perPageHeader: "100"},
		nil,
	)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
}

func (s *StubServerTestSuite) TestSendRequestWithBody() {
	createUserRequest := models.EndpointRequest{
		Path:               "/api/users",
		HTTPMethod:         http.MethodPost,
		ResponseHeaders:    map[string]string{contentTypeHeader: jsonContentType},
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
		map[string]string{contentTypeHeader: jsonContentType},
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
