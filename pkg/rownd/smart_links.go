package rownd

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type SmartLinkOptions struct {
    Email       string                 `json:"email,omitempty"`
    Phone       string                 `json:"phone,omitempty"`
    RedirectURL string                 `json:"redirect_url"`
    Data        map[string]interface{} `json:"data,omitempty"`
}

type SmartLink struct {
    Link      string `json:"link"`
    AppUserID string `json:"app_user_id"`
}

func (c *Client) CreateSmartLink(ctx context.Context, opts *SmartLinkOptions) (*SmartLink, error) {
    payload, err := json.Marshal(opts)
    if err != nil {
        return nil, NewError(ErrValidation, "failed to marshal request", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", 
        fmt.Sprintf("%s/hub/auth/magic", c.BaseURL), 
        bytes.NewBuffer(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    var link SmartLink
    if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &link, nil
}