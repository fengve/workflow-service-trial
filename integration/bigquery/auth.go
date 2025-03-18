package bigquery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GetAccessTokenRequest struct {
	GrantType string `json:"grant_type"` // access token grant type.
	Assertion string `json:"assertion"`  // signed signature
}

type GetAccessTokenResponse struct {
	AccessToken string `json:"access_token"` // access token value.
	ExpiresIn   int    `json:"expires_in"`   // ttl in seconds.
	TokenType   string `json:"token_type"`   // Bearer
}

// Generates jwt token with service account email and private key.
// This signature is used as an assertion to get short-term access token.
func GenerateSignature(
	serviceAccountEmail string, serviceAccountPrivateKey string) (*string, error) {
	now := time.Now().Unix()

	// Parse private key
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(serviceAccountPrivateKey))
	if err != nil {
		return nil, err
	}

	payload := jwt.MapClaims{
		"iss":   serviceAccountEmail,
		"sub":   serviceAccountEmail,
		"aud":   "https://www.googleapis.com/oauth2/v4/token",
		"iat":   now,
		"exp":   now + 3600,
		"scope": "https://www.googleapis.com/auth/bigquery",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, payload)
	if err != nil {
		return nil, err
	}

	signature, err := token.SignedString(key)
	if err != nil {
		return nil, err
	}
	return &signature, err
}

// Exchange service account private key to the access token using GCP Auth API.
func GetAccessToken(serviceAccountEmail string, serviceAccountPrivateKey string) (*string, error) {
	// Generate signature for access token exchange.
	signature, err := GenerateSignature(serviceAccountEmail, serviceAccountPrivateKey)
	if err != nil {
		return nil, err
	}

	// Do http request for access token exchange
	getAccessTokenRequest := &GetAccessTokenRequest{
		GrantType: "urn:ietf:params:oauth:grant-type:jwt-bearer",
		Assertion: *signature,
	}
	getAccessTokenRequestJson, err := json.Marshal(getAccessTokenRequest)
	if err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequest(
		"POST", "https://oauth2.googleapis.com/token", bytes.NewBuffer(getAccessTokenRequestJson))
	if err != nil {
		return nil, err
	}
	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	// Check the response status code.
	if httpResponse.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get access token, status code: %d", httpResponse.StatusCode)
	}

	// Parse access token in response body.
	body := GetAccessTokenResponse{}
	err = json.NewDecoder(httpResponse.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return &body.AccessToken, nil
}
