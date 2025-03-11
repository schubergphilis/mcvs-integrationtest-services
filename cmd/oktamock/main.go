// Package main provides a mocked Okta server which can be used to create and validate JWT tokens.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/schubergphilis/mcvs-integrationtest-services/internal/oktamock/models"
	log "github.com/sirupsen/logrus"
)

// ErrUnsupportedSigningMethod represents an error when an unsupported signing method is provided.
type UnsupportedSigningMethodError struct {
	ProvidedMethod string
}

func (e UnsupportedSigningMethodError) Error() string {
	return fmt.Sprintf("unsupported signing method: %s", e.ProvidedMethod)
}

// SigningMethod represents the signing method for a JWT.
type SigningMethod struct {
	actualMethod *jwt.SigningMethodRSA
}

// Alg returns the algorithm as string.
func (s SigningMethod) Alg() string {
	return s.actualMethod.Alg()
}

// UnmarshalText marshals the signing method to text.
func (s *SigningMethod) UnmarshalText(text []byte) error {
	switch string(text) {
	case "RS256":
		s.actualMethod = jwt.SigningMethodRS256

		return nil
	case "RS384":
		s.actualMethod = jwt.SigningMethodRS384

		return nil
	case "RS512":
		s.actualMethod = jwt.SigningMethodRS512

		return nil
	}

	return UnsupportedSigningMethodError{
		ProvidedMethod: string(text),
	}
}

// Config represents the configuration.
type Config struct {
	ServerConfig ServerConfig
	JWTConfig    JWTConfig
}

// ServerConfig represents the server configuration.
type ServerConfig struct {
	Port int `env:"PORT" envDefault:"8080"`
}

// JWTConfig represents the JWT configuration.
type JWTConfig struct {
	Aud           string        `env:"AUD"            envDefault:"api://default"`
	Expiration    time.Duration `env:"EXPIRATION"     envDefault:"24h"`
	Groups        []string      `env:"GROUPS"         envDefault:""`
	Issuer        string        `env:"ISSUER"         envDefault:"http://localhost:8080"`
	KID           string        `env:"KID"            envDefault:"mock-kid"`
	SigningMethod SigningMethod `env:"SIGNING_METHOD" envDefault:"RS256"`
	Sub           string        `env:"SUB"            envDefault:""`
}

// NewConfig returns the config.
func NewConfig() (*Config, error) {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func main() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	oktaMockServer, err := NewOktaMockServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/.well-known/openid-configuration", oktaMockServer.handleOpenIDConfig)
	http.HandleFunc("/v1/keys", oktaMockServer.handleGetJWKS)
	http.HandleFunc("/token", oktaMockServer.handleGetValidJWT)

	//nolint: gosec
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServerConfig.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// OktaMockServer represents a mock Okta server which can be used to create and validate JWT tokens.
// Serves as a subtitute for using an actual Okta Server.
type OktaMockServer struct {
	audience, issuer, sub string
	expiration            time.Duration
	groups                []string

	privKey *rsa.PrivateKey
	jwkKey  jwk.Key
}

// CustomClaimsRequest represents the JSON structure for requests that include custom claims for JWT tokens.
type CustomClaimsRequest struct {
	CustomClaims map[string]interface{} `json:"custom_claims"`
}

func (o *OktaMockServer) handleGetValidJWT(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var claimsReq CustomClaimsRequest
	if err := decoder.Decode(&claimsReq); err != nil {
		http.Error(w, "Okta mock expects custom claims to be present in token request", http.StatusBadRequest)

		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"aud":    o.audience,
		"exp":    now.Add(o.expiration).Unix(),
		"Groups": o.groups,
		"iat":    now.Unix(),
		"iss":    o.issuer,
		"nbf":    now.AddDate(0, 0, -1).Unix(),
		"sub":    o.sub,
	}

	// Add custom claims
	for key, value := range claimsReq.CustomClaims {
		claims[key] = value
	}

	// Create a new token with these claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = o.jwkKey.KeyID()

	res, err := token.SignedString(o.privKey)
	if err != nil {
		log.WithError(err).Error("unable to generate the signed JWT string")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	log.WithFields(log.Fields{"jwt": res}).Info("generated")

	// Prepare and send the response.
	tokenResponse := models.ValidJWTResponse{
		AccessToken: res,
	}
	b, err := json.Marshal(tokenResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).WithFields(log.Fields{"httpStatusCode": http.StatusInternalServerError}).Error("unable to generate the signed JWT string")

		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.WithError(err).Error("unable to write token response")
	}
}

func (o *OktaMockServer) handleGetJWKS(w http.ResponseWriter, _ *http.Request) {
	resp := models.JWKSResponse{
		Keys: []jwk.Key{o.jwkKey},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).WithFields(log.Fields{"httpStatusCode": http.StatusInternalServerError}).Error("unable to handle get JWKS")

		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.WithError(err).Error("unable to write JWKS")
	}
}

func (o *OktaMockServer) handleOpenIDConfig(w http.ResponseWriter, _ *http.Request) {
	resp := models.OpenIDConfigurationResponse{
		JwksURI: fmt.Sprintf("%s/v1/keys", o.issuer),
	}
	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithError(err).WithFields(log.Fields{"httpStatusCode": http.StatusInternalServerError}).Error("unable to handle openID config")

		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.WithError(err).Error("unable to write openID config")
	}
}

// NewOktaMockServer returns a new OktaMockServer.
func NewOktaMockServer(cfg *Config) (*OktaMockServer, error) {
	privKeyRSA, jwkKey, err := genRSAKeyAndJWK(&cfg.JWTConfig)
	if err != nil {
		return nil, err
	}

	return &OktaMockServer{
		audience:   cfg.JWTConfig.Aud,
		expiration: cfg.JWTConfig.Expiration,
		groups:     cfg.JWTConfig.Groups,
		issuer:     cfg.JWTConfig.Issuer,
		jwkKey:     jwkKey,
		privKey:    privKeyRSA,
		sub:        cfg.JWTConfig.Sub,
	}, nil
}

func genRSAKeyAndJWK(cfg *JWTConfig) (*rsa.PrivateKey, jwk.Key, error) {
	bitSize := 4096

	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, nil, err
	}

	err = privateKey.Validate()
	if err != nil {
		return nil, nil, err
	}

	jwkKey, err := jwk.PublicKeyOf(privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	err = jwkKey.Set(jwk.KeyIDKey, cfg.KID)
	if err != nil {
		return nil, nil, err
	}
	err = jwkKey.Set(jwk.AlgorithmKey, cfg.SigningMethod.Alg())
	if err != nil {
		return nil, nil, err
	}

	return privateKey, jwkKey, nil
}
