package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rgthelen/rownd-go-test/pkg/rownd"
)

// AuthTokens represents the tokens returned from auth operations
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
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

// InitiateAuth starts the authentication process
func InitiateAuth(ctx context.Context, client *rownd.Client, req *rownd.AuthInitRequest) (*rownd.AuthInitResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, rownd.NewError(rownd.ErrValidation, "failed to marshal request", err)
	}

	apiReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/hub/auth/init", client.BaseURL),
		bytes.NewBuffer(payload))
	if err != nil {
		return nil, rownd.NewError(rownd.ErrAPI, "failed to create request", err)
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("X-Rownd-App-Key", client.AppKey)

	resp, err := client.HTTPClient.Do(apiReq)
	if err != nil {
		return nil, rownd.NewError(rownd.ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	var initResp rownd.AuthInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return nil, rownd.NewError(rownd.ErrAPI, "failed to decode response", err)
	}

	return &initResp, nil
}

// CompleteAuth completes the authentication process
func CompleteAuth(ctx context.Context, client *rownd.Client, req *rownd.AuthCompleteRequest) (*rownd.AuthCompleteResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, rownd.NewError(rownd.ErrValidation, "failed to marshal request", err)
	}

	apiReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/hub/auth/complete", client.BaseURL),
		bytes.NewBuffer(payload))
	if err != nil {
		return nil, rownd.NewError(rownd.ErrAPI, "failed to create request", err)
	}

	apiReq.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(apiReq)
	if err != nil {
		return nil, rownd.NewError(rownd.ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	var completeResp rownd.AuthCompleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&completeResp); err != nil {
		return nil, rownd.NewError(rownd.ErrAPI, "failed to decode response", err)
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
	endpoint := fmt.Sprintf("%s/hub/auth/magic/%s", client.BaseURL, linkID)
	fmt.Printf("\nAttempting to redeem magic link at: %s\n", endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, rownd.NewError(rownd.ErrAPI, "failed to create request", err)
	}

	req.Header.Set("User-Agent", "rownd sdk")
	fmt.Printf("Request headers: %+v\n", req.Header)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, rownd.NewError(rownd.ErrNetwork, fmt.Sprintf("request failed: %v", err), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, rownd.NewError(rownd.ErrAPI, "failed to read response body", err)
	}
	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, rownd.NewError(rownd.ErrAPI, fmt.Sprintf("failed to redeem magic link: %s", string(body)), nil)
	}

	var magicLinkResp MagicLinkResponse
	if err := json.Unmarshal(body, &magicLinkResp); err != nil {
		return nil, rownd.NewError(rownd.ErrAPI, "failed to decode response", err)
	}

	fmt.Printf("Successfully redeemed magic link. Access token length: %d\n", len(magicLinkResp.AccessToken))
	return &magicLinkResp, nil
}