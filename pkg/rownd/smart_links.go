package rownd

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type SmartLinkOptions struct {
    Email            string                 `json:"email,omitempty"`
    Phone            string                 `json:"phone,omitempty"`
    RedirectURL      string                 `json:"redirect_url"`
    Data             map[string]interface{} `json:"data,omitempty"`
    PostRedirectURL  string                 `json:"post_redirect_url,omitempty"`
    AppID            string                 `json:"app_id,omitempty"`
    Type             string                 `json:"type,omitempty"`            // "magic_link" or "verification"
    Intent           string                 `json:"intent,omitempty"`          // "sign_in" or "sign_up"
    ExpiresIn        int                    `json:"expires_in,omitempty"`      // seconds
    AutoSignIn       bool                   `json:"auto_sign_in,omitempty"`
    Purpose          string                 `json:"purpose,omitempty"`         // "auth"
    VerificationType string                 `json:"verification_type,omitempty"` // "email" or "phone"
    UserID           string                 `json:"user_id,omitempty"`
    Expiration       string                 `json:"expiration,omitempty"`      // e.g. "30d"
    GroupToJoin      string                 `json:"group_to_join,omitempty"`
}

type SmartLink struct {
    Link           string `json:"link"`
    AppUserID      string `json:"app_user_id"`
    ExpiresAt      string `json:"expires_at,omitempty"`
    VerificationID string `json:"verification_id,omitempty"`
}

func (c *Client) CreateSmartLink(ctx context.Context, opts *SmartLinkOptions) (*SmartLink, error) {
    if opts.RedirectURL == "" {
        return nil, NewError(ErrValidation, "redirect_url is required", nil)
    }

    if opts.Email == "" && opts.Phone == "" {
        return nil, NewError(ErrValidation, "either email or phone is required", nil)
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