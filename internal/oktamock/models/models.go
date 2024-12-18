package models

import "github.com/lestrrat-go/jwx/v2/jwk"

// OpenIDConfigurationResponse represents the response from the OpenID Configuration endpoint.
type OpenIDConfigurationResponse struct {
	Issuer                                                    string   `json:"issuer"`
	AuthorizationEndpoint                                     string   `json:"authorization_endpoint"`
	TokenEndpoint                                             string   `json:"token_endpoint"`
	UserinfoEndpoint                                          string   `json:"userinfo_endpoint"`
	RegistrationEndpoint                                      string   `json:"registration_endpoint"`
	JwksURI                                                   string   `json:"jwks_uri"`
	ResponseTypesSupported                                    []string `json:"response_types_supported"`
	ResponseModesSupported                                    []string `json:"response_modes_supported"`
	GrantTypesSupported                                       []string `json:"grant_types_supported"`
	SubjectTypesSupported                                     []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported                          []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                                           []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported                         []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                                           []string `json:"claims_supported"`
	CodeChallengeMethodsSupported                             []string `json:"code_challenge_methods_supported"`
	IntrospectionEndpoint                                     string   `json:"introspection_endpoint"`
	IntrospectionEndpointAuthMethodsSupported                 []string `json:"introspection_endpoint_auth_methods_supported"`
	RevocationEndpoint                                        string   `json:"revocation_endpoint"`
	RevocationEndpointAuthMethodsSupported                    []string `json:"revocation_endpoint_auth_methods_supported"`
	EndSessionEndpoint                                        string   `json:"end_session_endpoint"`
	RequestParameterSupported                                 bool     `json:"request_parameter_supported"`
	RequestObjectSigningAlgValuesSupported                    []string `json:"request_object_signing_alg_values_supported"`
	DeviceAuthorizationEndpoint                               string   `json:"device_authorization_endpoint"`
	PushedAuthorizationRequestEndpoint                        string   `json:"pushed_authorization_request_endpoint"`
	BackchannelTokenDeliveryModesSupported                    []string `json:"backchannel_token_delivery_modes_supported"`
	BackchannelAuthenticationRequestSigningAlgValuesSupported []string `json:"backchannel_authentication_request_signing_alg_values_supported"`
	DpopSigningAlgValuesSupported                             []string `json:"dpop_signing_alg_values_supported"`
}

// JWKSResponse represents the response from the JWKS endpoint.
type JWKSResponse struct {
	Keys []jwk.Key `json:"keys"`
}

// ValidJWTResponse represents the response from the valid JWT endpoint.
type ValidJWTResponse struct {
	AccessToken string `json:"access_token"`
}
