package stubserver

import (
	"fmt"
	"strings"
	"sync"
)

// ResponseManager manages the stub server responses.
type ResponseManager struct {
	mu        sync.RWMutex
	endpoints map[string]EndpointConfiguration
}

// NewResponseManager creates a new instance of ResponseManager.
func NewResponseManager() *ResponseManager {
	return &ResponseManager{
		endpoints: make(map[string]EndpointConfiguration),
	}
}

// EndpointID represents a unique identifier for an endpoint.
type EndpointID struct {
	Path               string
	HTTPMethod         string
	QueryParamsToMatch map[string]string
	HeadersToMatch     map[string]string
}

// EndpointConfiguration represents the configuration for a stub endpoint.
type EndpointConfiguration struct {
	EndpointID         EndpointID
	ResponseHeaders    map[string]string
	ResponseBody       string
	ResponseStatusCode int
}

// GetID generates a unique ID for the endpoint based on its path, method, headers, and query parameters.
func GetID(ei *EndpointID) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s:%s", ei.Path, ei.HTTPMethod))
	for nameOfHeaderToMatch, valueOfHeaderToMatch := range ei.HeadersToMatch {
		builder.WriteString(fmt.Sprintf(":%s=%s", nameOfHeaderToMatch, valueOfHeaderToMatch))
	}
	for nameOfQueryParamToMatch, valueOfQueryParamToMatch := range ei.QueryParamsToMatch {
		builder.WriteString(fmt.Sprintf(":%s=%s", nameOfQueryParamToMatch, valueOfQueryParamToMatch))
	}

	return builder.String()
}

// ValidateEndpoint valid endpoint configuration.
func ValidateEndpoint(ep EndpointConfiguration) error {
	if ep.EndpointID.Path == "" || ep.EndpointID.HTTPMethod == "" {
		return fmt.Errorf("path and method are required")
	}
	if ep.ResponseBody == "" {
		return fmt.Errorf("response body is required")
	}

	err := validateHTTPMethods(ep)
	if err != nil {
		return err
	}

	return nil
}

func validateHTTPMethods(ep EndpointConfiguration) error {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	isValidMethod := false
	for _, method := range validMethods {
		if ep.EndpointID.HTTPMethod == method {
			isValidMethod = true

			break
		}
	}
	if !isValidMethod {
		return fmt.Errorf("invalid HTTP method: %s", ep.EndpointID.HTTPMethod)
	}

	return nil
}

// AddEndpoint adds a new endpoint configuration to the manager.
func (rm *ResponseManager) AddEndpoint(ec EndpointConfiguration) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	err := ValidateEndpoint(ec)
	if err != nil {
		return err
	}

	endpointID := GetID(&ec.EndpointID)
	if _, exists := rm.endpoints[endpointID]; exists {
		return fmt.Errorf("endpoint already exists: %s", endpointID)
	}
	rm.endpoints[endpointID] = ec

	return nil
}

// GetEndpointByEndpointID retrieves the configuration for a given endpoint.
func (rm *ResponseManager) GetEndpointByEndpointID(ei *EndpointID) (EndpointConfiguration, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	endpointID := GetID(ei)
	config, exists := rm.endpoints[endpointID]
	if !exists {
		return EndpointConfiguration{}, fmt.Errorf("endpoint not found: %s", ei)
	}

	return config, nil
}

// GetAllEndpointConfigurations retrieves all endpoint configurations.
func (rm *ResponseManager) GetAllEndpointConfigurations() []EndpointConfiguration {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	configs := make([]EndpointConfiguration, 0, len(rm.endpoints))
	for _, config := range rm.endpoints {
		configs = append(configs, config)
	}

	return configs
}

// DeleteEndpointByEndpointID deletes an endpoint configuration by its ID.
func (rm *ResponseManager) DeleteEndpointByEndpointID(ei *EndpointID) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	endpointID := GetID(ei)
	if _, exists := rm.endpoints[endpointID]; !exists {
		return fmt.Errorf("endpoint not found: %s", endpointID)
	}
	delete(rm.endpoints, endpointID)

	return nil
}

// DeleteEndpointByPath deletes all endpoint configurations with the specified path.
func (rm *ResponseManager) DeleteEndpointByPath(path string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for endpointID, endpoint := range rm.endpoints {
		if endpoint.EndpointID.Path == path {
			delete(rm.endpoints, endpointID)
		}
	}

	return nil
}

// DeleteEndpointByPathAndMethod deletes all endpoint configurations with the specified path and method.
func (rm *ResponseManager) DeleteEndpointByPathAndMethod(path, method string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for endpointID, endpoint := range rm.endpoints {
		if endpoint.EndpointID.Path == path && endpoint.EndpointID.HTTPMethod == method {
			delete(rm.endpoints, endpointID)
		}
	}

	return nil
}

// DeleteAllEndpoints deletes all endpoint configurations.
func (rm *ResponseManager) DeleteAllEndpoints() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for endpointID := range rm.endpoints {
		delete(rm.endpoints, endpointID)
	}
}
