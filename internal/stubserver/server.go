package stubserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	baseURLPath       = "/stubserver"
	healthEndpoint    = "/health"
	responsesEndpoint = "/responses"
)

const (
	maxBodySizeBytes = 1024 * 10
)

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

// Server represents the Gin HTTP server with router and handler.
type Server struct {
	Router          *gin.Engine
	responseManager *ResponseManager
}

// NewServer creates a new instance of Server with configured routes.
func NewServer() *Server {
	router := gin.Default()
	responseManager := NewResponseManager()

	server := &Server{
		Router:          router,
		responseManager: responseManager,
	}

	router.GET(healthEndpoint, server.health)
	router.POST(baseURLPath+responsesEndpoint, server.addResponse)
	router.GET(baseURLPath+responsesEndpoint, server.getAllResponses)
	router.DELETE(baseURLPath+responsesEndpoint, server.deleteAllResponses)

	router.NoRoute(server.catchAll)

	return server
}

func (s *Server) health(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (s *Server) addResponse(c *gin.Context) {
	var request EndpointRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})

		return
	}

	if request.Path == "" || request.HTTPMethod == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Path and HTTP method are required"})

		return
	}

	if request.ResponseBody == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Response body is required"})

		return
	}

	endpointConfig := EndpointConfiguration{
		EndpointID: EndpointID{
			Path:               request.Path,
			HTTPMethod:         request.HTTPMethod,
			QueryParamsToMatch: request.QueryParamsToMatch,
			HeadersToMatch:     request.HeadersToMatch,
		},
		ResponseHeaders:    request.ResponseHeaders,
		ResponseBody:       request.ResponseBody,
		ResponseStatusCode: request.ResponseStatusCode,
	}

	err := s.responseManager.AddEndpoint(endpointConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) getAllResponses(c *gin.Context) {
	configs := s.responseManager.GetAllEndpointConfigurations()

	responses := make([]EndpointResponse, 0, len(configs))
	for _, config := range configs {
		responses = append(responses, EndpointResponse{
			Path:               config.EndpointID.Path,
			HTTPMethod:         config.EndpointID.HTTPMethod,
			QueryParamsToMatch: config.EndpointID.QueryParamsToMatch,
			HeadersToMatch:     config.EndpointID.HeadersToMatch,
			ResponseHeaders:    config.ResponseHeaders,
			ResponseBody:       config.ResponseBody,
			ResponseStatusCode: config.ResponseStatusCode,
		})
	}

	c.JSON(http.StatusOK, EndpointListResponse{Endpoints: responses})
}

func (s *Server) deleteAllResponses(c *gin.Context) {
	s.responseManager.DeleteAllEndpoints()
	c.Status(http.StatusOK)
}

func (s *Server) catchAll(c *gin.Context) {
	logRequestContext(c)

	endpointID := EndpointID{
		Path:               c.Request.URL.Path,
		HTTPMethod:         c.Request.Method,
		QueryParamsToMatch: flattenQueryParams(c),
		HeadersToMatch:     flattenHeaders(c),
	}

	config, err := s.responseManager.GetEndpointByEndpointID(&endpointID)
	if err != nil {
		log.WithFields(log.Fields{"urlPath": c.Request.URL.Path}).Error("endpoint not found")
		c.Status(http.StatusNotFound)

		return
	}

	// 1. Set response headers
	for key, value := range config.ResponseHeaders {
		c.Header(key, value)
	}

	// 2. Set status code (default to 200 if not set)
	statusCode := config.ResponseStatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	c.String(statusCode, config.ResponseBody)
}

func flattenQueryParams(c *gin.Context) map[string]string {
	result := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}

func flattenHeaders(c *gin.Context) map[string]string {
	result := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}

func logRequestContext(c *gin.Context) {
	var requestInfo strings.Builder

	requestInfo.WriteString(fmt.Sprintf("Request Method: %s\n", c.Request.Method))
	requestInfo.WriteString(fmt.Sprintf("Absolute URL: %s\n", c.Request.URL.String()))
	requestInfo.WriteString(fmt.Sprintf("Absolute Path: %s\n", c.Request.URL.Path))
	requestInfo.WriteString(fmt.Sprintf("Host: %s\n", c.Request.Host))
	requestInfo.WriteString(fmt.Sprintf("Remote Address: %s\n", c.Request.RemoteAddr))

	requestInfo.WriteString("Headers:\n")
	for name, values := range c.Request.Header {
		if strings.EqualFold(name, "Authorization") {
			requestInfo.WriteString(fmt.Sprintf("  %s: *****\n", name))

			continue
		}
		for _, value := range values {
			requestInfo.WriteString(fmt.Sprintf("  %s: %s\n", name, value))
		}
	}

	requestInfo.WriteString("Query Parameters:\n")
	for name, values := range c.Request.URL.Query() {
		for _, value := range values {
			requestInfo.WriteString(fmt.Sprintf("  %s: %s\n", name, value))
		}
	}

	body, err := c.GetRawData()
	if err == nil && len(body) > 0 {
		if len(body) > maxBodySizeBytes {
			body = body[:maxBodySizeBytes]
		}
		requestInfo.WriteString(fmt.Sprintf("Body Content:\n%s\n", string(body)))

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	log.WithFields(log.Fields{"context": requestInfo.String()}).Info("log request")
}
