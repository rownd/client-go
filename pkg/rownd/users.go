package rownd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "context"
    "strings"
)

func (c *Client) GetUser(ctx context.Context, userID string, tokenInfo *TokenValidationResponse) (*User, error) {
    // Get app ID from token claims
    var appID string
    if tokenInfo != nil {
        if aud, ok := tokenInfo.DecodedToken["aud"].([]string); ok && len(aud) > 0 {
            if strings.HasPrefix(aud[0], "app:") {
                appID = aud[0][4:]
            }
        }
    }

    // If no app ID in token, use the one from client config
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