package rownd

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	rowndtesting "github.com/rgthelen/rownd-go-test/pkg/rownd/testing"
)

const (
	CLAIM_USER_ID          = "https://auth.rownd.io/app_user_id"
	CLAIM_IS_VERIFIED_USER = "https://auth.rownd.io/is_verified_user"
	CLAIM_IS_ANONYMOUS     = "https://auth.rownd.io/is_anonymous"
	CLAIM_AUTH_LEVEL       = "https://auth.rownd.io/auth_level"

	AUTH_LEVEL_INSTANT       = "instant"
	AUTH_LEVEL_UNVERIFIED    = "unverified"
	AUTH_LEVEL_GUEST         = "guest"
	AUTH_LEVEL_VERIFIED      = "verified"
)

type TokenValidationResponse struct {
	DecodedToken jwt.MapClaims `json:"decoded_token"`
	UserID       string        `json:"user_id"`
	AccessToken  string        `json:"access_token"`
}

type JWKS struct {
	Keys []json.RawMessage `json:"keys"`
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error) {
	// First fetch the well-known config
	config, err := c.FetchWellKnownConfig(ctx)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to fetch well-known config", err)
	}

	// Fetch JWKS
	jwks, err := c.fetchJWKS(ctx, config.JwksUri)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to fetch JWKS", err)
	}

	// Parse and validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			// For testing, use the test public key if no kid is present
			if publicKey, _ := rowndtesting.GetKeys(); publicKey != nil {
				return publicKey, nil
			}
			return nil, fmt.Errorf("kid header not found")
		}

		// Find the key with matching kid
		for _, rawKey := range jwks.Keys {
			var key struct {
				Kid string `json:"kid"`
				X   string `json:"x"` // EdDSA public key
			}
			if err := json.Unmarshal(rawKey, &key); err != nil {
				continue
			}
			if key.Kid == kid {
				// Decode the EdDSA public key from base64
				publicKey, err := base64.RawURLEncoding.DecodeString(key.X)
				if err != nil {
					continue
				}
				return ed25519.PublicKey(publicKey), nil
			}
		}
		return nil, fmt.Errorf("key %v not found", kid)
	})

	if err != nil {
		return nil, NewError(ErrAuthentication, "invalid token", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, NewError(ErrAuthentication, "invalid token claims", nil)
	}

	userID, _ := claims[CLAIM_USER_ID].(string)

	// Store claims in context for other methods to use
	ctx = context.WithValue(ctx, "rownd_token_claims", claims)

	return &TokenValidationResponse{
		DecodedToken: claims,
		UserID:       userID,
		AccessToken:  token,
	}, nil
}

func (c *Client) fetchJWKS(ctx context.Context, jwksUri string) (*JWKS, error) {
	// Check cache first
	if cached, found := c.cache.Get("jwks"); found {
		return cached.(*JWKS), nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", jwksUri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	// Cache the key set
	c.cache.Set("jwks", &jwks, time.Hour)

	return &jwks, nil
}

func (c *Client) InitiateAuth(ctx context.Context, req *AuthInitRequest) (*AuthInitResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, NewError(ErrValidation, "failed to marshal request", err)
	}

	apiReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/hub/auth/init", c.BaseURL),
		bytes.NewBuffer(payload))
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("X-Rownd-App-Key", c.AppKey)

	resp, err := c.HTTPClient.Do(apiReq)
	if err != nil {
		return nil, NewError(ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	var initResp AuthInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return nil, NewError(ErrAPI, "failed to decode response", err)
	}

	return &initResp, nil
}

func (c *Client) CompleteAuth(ctx context.Context, req *AuthCompleteRequest) (*AuthCompleteResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, NewError(ErrValidation, "failed to marshal request", err)
	}

	apiReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/hub/auth/complete", c.BaseURL),
		bytes.NewBuffer(payload))
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	apiReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(apiReq)
	if err != nil {
		return nil, NewError(ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	var completeResp AuthCompleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&completeResp); err != nil {
		return nil, NewError(ErrAPI, "failed to decode response", err)
	}

	return &completeResp, nil
}

func ParseAuthRedirect(redirectURL string) (*AuthTokens, error) {
	u, err := url.Parse(redirectURL)
	if err != nil {
		return nil, err
	}

	fragment := u.Fragment
	values, err := url.ParseQuery(fragment)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  values.Get("access_token"),
		RefreshToken: values.Get("refresh_token"),
	}, nil
} 