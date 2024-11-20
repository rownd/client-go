package rownd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
)

const (
	CLAIM_USER_ID = "https://auth.rownd.io/app_user_id"
	CLAIM_IS_VERIFIED_USER = "https://auth.rownd.io/is_verified_user"
)

type TokenValidationResponse struct {
	DecodedToken jwt.MapClaims `json:"decoded_token"`
	UserID       string        `json:"user_id"`
	AccessToken  string        `json:"access_token"`
}

func (c *Client) ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error) {
	// First fetch the well-known config
	config, err := c.FetchWellKnownConfig(ctx)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to fetch well-known config", err)
	}

	// Fetch JWKS
	keySet, err := c.fetchJWKS(ctx, config.JwksUri)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to fetch JWKS", err)
	}

	// Parse and validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid header not found")
		}

		key := keySet.LookupKeyID(kid)
		if key == nil {
			return nil, fmt.Errorf("key %v not found", kid)
		}

		return key.PublicKey, nil
	})

	if err != nil {
		return nil, NewError(ErrAuthentication, "invalid token", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, NewError(ErrAuthentication, "invalid token claims", nil)
	}

	// Extract user ID from claims
	userID, _ := claims[CLAIM_USER_ID].(string)

	return &TokenValidationResponse{
		DecodedToken: claims,
		UserID:       userID,
		AccessToken:  token,
	}, nil
}

func (c *Client) fetchJWKS(ctx context.Context, jwksUri string) (*jwt.KeySet, error) {
	// Check cache first
	if cachedKeySet := c.cache.Get("jwks"); cachedKeySet != nil {
		return cachedKeySet.(*jwt.KeySet), nil
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

	var jwks struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}

	keySet, err := jwt.ParseKeySet(jwks.Keys)
	if err != nil {
		return nil, err
	}

	// Cache the key set
	c.cache.Set("jwks", keySet, time.Hour)

	return keySet, nil
} 