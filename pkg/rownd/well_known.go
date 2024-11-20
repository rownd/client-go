package rownd

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

func (c *Client) FetchWellKnownConfig(ctx context.Context) (*WellKnownConfig, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", 
        fmt.Sprintf("%s/hub/auth/.well-known/oauth-authorization-server", c.BaseURL), 
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    var config WellKnownConfig
    if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &config, nil
}