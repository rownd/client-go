package rownd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "context"
    "github.com/golang-jwt/jwt"
)

func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
    // Get app ID from token claims if available
    var appID string
    if tokenInfo, ok := ctx.Value("rownd_token_info").(*TokenValidationResponse); ok {
        if aud, ok := tokenInfo.DecodedToken["aud"].([]interface{}); ok && len(aud) > 0 {
            if audStr, ok := aud[0].(string); ok {
                if len(audStr) > 4 && audStr[:4] == "app:" {
                    appID = audStr[4:]
                }
            }
        }
    }

    // If no app ID in context, use the one from client config
    if appID == "" {
        appID = c.AppID
    }

    if appID == "" {
        return nil, fmt.Errorf("app ID not found in token or client config")
    }

    req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/applications/%s/users/%s/data", c.BaseURL, appID, userID), nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        var apiErr APIResponse
        if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
            return nil, fmt.Errorf("failed to decode error response: %w", err)
        }
        return nil, fmt.Errorf("API error: %s", apiErr.Error)
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("failed to decode user response: %w", err)
    }

    return &user, nil
}

func (c *Client) UpdateUser(userID string, data map[string]interface{}) (*User, error) {
    payload, err := json.Marshal(map[string]interface{}{
        "data": data,
    })
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/hub/users/%s", c.BaseURL, userID), bytes.NewBuffer(payload))
    if err != nil {
        return nil, err
    }

    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }

    return &user, nil
}