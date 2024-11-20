package rownd

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TokenValidationResponse struct {
	Valid    bool   `json:"valid"`
	UserID   string `json:"user_id,omitempty"`
	AppID    string `json:"app_id,omitempty"`
	IssuedAt int64  `json:"iat,omitempty"`
	ExpireAt int64  `json:"exp,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (c *Client) ValidateToken(token string) (*TokenValidationResponse, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/hub/auth/validate", c.BaseURL), nil)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, NewError(ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	var result TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewError(ErrAPI, "failed to decode response", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &result, NewError(ErrAuthentication, result.Error, nil)
	}

	if !result.Valid {
		return &result, NewError(ErrAuthentication, "token is invalid", nil)
	}

	return &result, nil
} 