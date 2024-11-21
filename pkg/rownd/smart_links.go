package rownd

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type SmartLinkOptions struct {
    Purpose          string                 `json:"purpose"`          // "auth"
    VerificationType string                 `json:"verification_type"` // "email" or "phone"
    Data             map[string]interface{} `json:"data"`
    RedirectURL      string                 `json:"redirect_url"`
    UserID           string                 `json:"user_id,omitempty"`
    Expiration       string                 `json:"expiration,omitempty"`      // e.g. "30d"
    GroupToJoin      string                 `json:"group_to_join,omitempty"`
}

type SmartLink struct {
    Link      string `json:"link"`
    AppUserID string `json:"app_user_id"`
}

func (c *Client) CreateSmartLink(ctx context.Context, opts *SmartLinkOptions) (*SmartLink, error) {
    if opts.RedirectURL == "" {
        return nil, NewError(ErrValidation, "redirect_url is required", nil)
    }

    if opts.Purpose == "" {
        return nil, NewError(ErrValidation, "purpose is required", nil)
    }

    if opts.VerificationType == "" {
        return nil, NewError(ErrValidation, "verification_type is required", nil)
    }

    if opts.Data == nil || (opts.VerificationType == "email" && opts.Data["email"] == "") {
        return nil, NewError(ErrValidation, "data.email is required for email verification", nil)
    }

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

    if resp.StatusCode != http.StatusOK {
        var apiErr APIResponse
        if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
            return nil, NewError(ErrAPI, "failed to decode error response", err)
        }
        return nil, NewError(ErrAPI, apiErr.Error, nil)
    }

    var link SmartLink
    if err := json.NewDecoder(resp.Body).Decode(&link); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &link, nil
}