package testutils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rgthelen/rownd-go-sdk/pkg/rownd"
)

// AuthTokens represents the tokens returned from auth operations
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

// AuthInitRequest represents the request to initialize authentication
type AuthInitRequest struct {
	Email string `json:"email"`
}

// AuthInitResponse represents the response from initializing authentication
type AuthInitResponse struct {
	Status string `json:"status"`
}

// AuthCompleteRequest represents the request to complete authentication
type AuthCompleteRequest struct {
	Code string `json:"code"`
}

// AuthCompleteResponse represents the response from completing authentication
type AuthCompleteResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// MagicLinkResponse represents the response from redeeming a magic link
type MagicLinkResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AppUserID    string `json:"app_user_id"`
	AppID        string `json:"app_id"`
	LastSignIn   string `json:"last_sign_in"`
	RedirectURL  string `json:"redirect_url"`
}

// TokenClaims represents the JWT claims for testing
type TokenClaims struct {
	AppUserID      string `json:"https://auth.rownd.io/app_user_id"`
	IsUserVerified bool   `json:"https://auth.rownd.io/is_verified_user"`
	AuthLevel      string `json:"https://auth.rownd.io/auth_level"`
}

// TokenInfo represents the token validation result for testing
type TokenInfo struct {
	Claims TokenClaims
	UserID string
}

// InitiateAuth starts the authentication process
func InitiateAuth(ctx context.Context, client *rownd.Client, req *AuthInitRequest) (*AuthInitResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/hub/auth/init", client.GetBaseURL())
	apiReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("X-Rownd-App-Key", client.GetAppKey())

	resp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var initResp AuthInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &initResp, nil
}

// CompleteAuth completes the authentication process
func CompleteAuth(ctx context.Context, client *rownd.Client, req *AuthCompleteRequest) (*AuthCompleteResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/hub/auth/complete", client.GetBaseURL())
	apiReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	apiReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var completeResp AuthCompleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&completeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &completeResp, nil
}

// ParseAuthRedirect parses the redirect URL to extract auth tokens
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

// RedeemMagicLink exchanges a magic link ID for authentication tokens
func RedeemMagicLink(ctx context.Context, client *rownd.Client, linkID string) (*MagicLinkResponse, error) {
	endpoint := fmt.Sprintf("%s/hub/auth/magic/%s", client.GetBaseURL(), linkID)
	fmt.Printf("\nAttempting to redeem magic link at: %s\n", endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "rownd sdk")
	fmt.Printf("Request headers: %+v\n", req.Header)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to redeem magic link: %s", string(body))
	}

	var magicLinkResp MagicLinkResponse
	if err := json.Unmarshal(body, &magicLinkResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Successfully redeemed magic link. Access token length: %d\n", len(magicLinkResp.AccessToken))
	return &magicLinkResp, nil
}

// ValidateTokenForTest is a simplified token validator for testing
func ValidateTokenForTest(ctx context.Context, client *rownd.Client, token string) (*TokenInfo, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &TokenInfo{
		Claims: claims,
		UserID: claims.AppUserID,
	}, nil
}
