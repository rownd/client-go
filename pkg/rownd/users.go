package rownd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "context"
)

func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/hub/users/%s", c.BaseURL, userID), nil)
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