package rownd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func (c *Client) GetUser(userID string) (*User, error) {
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/hub/users/%s", c.BaseURL, userID), nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

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