package stubserver

import (
	"fmt"
	"slices"
	"sort"
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

	headerKeys := make([]string, 0, len(ei.HeadersToMatch))
	for key := range ei.HeadersToMatch {
		headerKeys = append(headerKeys, key)
	}

	sort.Strings(headerKeys)

	for _, key := range headerKeys {
		builder.WriteString(fmt.Sprintf(":%s=%s", key, ei.HeadersToMatch[key]))
	}

	queryKeys := make([]string, 0, len(ei.QueryParamsToMatch))
	for key := range ei.QueryParamsToMatch {
		queryKeys = append(queryKeys, key)
	}

	sort.Strings(queryKeys)

	for _, key := range queryKeys {
		builder.WriteString(fmt.Sprintf(":%s=%s", key, ei.QueryParamsToMatch[key]))
	}

	return strings.ToLower(builder.String())
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

	isValidMethod := slices.Contains(validMethods, ep.EndpointID.HTTPMethod)

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

func (rm *ResponseManager) MatchEndpoint(ei *EndpointID) (*EndpointConfiguration, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	endpointsScoring := make(map[int]*EndpointConfiguration)
	maxScore := -1

	for _, endpoint := range rm.endpoints {
		if endpoint.EndpointID.Path == ei.Path && endpoint.EndpointID.HTTPMethod == ei.HTTPMethod {
			currentScore := calculateMatch(&endpoint, ei)

			if _, exists := endpointsScoring[currentScore]; exists {
				return nil, fmt.Errorf("can't match for request: %s, to many matches", GetID(&endpoint.EndpointID))
			}

			endpointsScoring[currentScore] = &endpoint

			if currentScore > maxScore {
				maxScore = currentScore
			}
		}
	}

	if maxScore == -1 {
		return nil, fmt.Errorf("no endpoints matched the given request: %s", ei.Path)
	}

	return endpointsScoring[maxScore], nil
}

func calculateMatch(ec *EndpointConfiguration, ei *EndpointID) int {
	counter := 0

	for eiQueryParamToMatchName, eiQueryParamToMatchValue := range ei.QueryParamsToMatch {
		for queryParamToMatchName, queryParamToMatchValue := range ec.EndpointID.QueryParamsToMatch {
			if queryParamToMatchName == eiQueryParamToMatchName && queryParamToMatchValue == eiQueryParamToMatchValue {
				counter++
			}
		}
	}

	for eiHeadersToMatchName, eiHeadersToMatchValue := range ei.HeadersToMatch {
		for headerToMatchName, headerToMatchValue := range ec.EndpointID.HeadersToMatch {
			if headerToMatchName == eiHeadersToMatchName && headerToMatchValue == eiHeadersToMatchValue {
				counter++
			}
		}
	}

	return counter
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
