package rownd

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TokenValidationResponse struct {
	Valid    bool   `json:"valid"`
	UserID   string `json:"user_id"`
	AppID    string `json:"app_id"`
	IssuedAt int64  `json:"iat"`
	ExpireAt int64  `json:"exp"`
}

func (c *Client) ValidateToken(token string) (*TokenValidationResponse, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/hub/auth/validate", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
} 