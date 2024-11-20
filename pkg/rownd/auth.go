package rownd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/hub/auth/validate", c.BaseURL), nil)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, NewError(ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Response from Rownd: %s", string(body))

	var result TokenValidationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, NewError(ErrAPI, fmt.Sprintf("failed to decode response: %s", string(body)), err)
	}

	if resp.StatusCode != http.StatusOK {
		return &result, NewError(ErrAuthentication, fmt.Sprintf("authentication failed: %s", result.Error), nil)
	}

	return &result, nil
} 