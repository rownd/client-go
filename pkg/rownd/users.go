package rownd

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "context"
    "strings"
    "io"
    "mime/multipart"
)

func (c *Client) GetUser(ctx context.Context, userID string, tokenInfo *TokenValidationResponse) (*User, error) {
    // Get app ID from token claims
    var appID string
    if tokenInfo != nil && tokenInfo.DecodedToken != nil {
        if aud, exists := tokenInfo.DecodedToken["aud"]; exists {
            switch v := aud.(type) {
            case []interface{}:
                if len(v) > 0 {
                    if audStr, ok := v[0].(string); ok && strings.HasPrefix(audStr, "app:") {
                        appID = audStr[4:]
                    }
                }
            case []string:
                if len(v) > 0 && strings.HasPrefix(v[0], "app:") {
                    appID = v[0][4:]
                }
            case string:
                if strings.HasPrefix(v, "app:") {
                    appID = v[4:]
                }
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

func (c *Client) UpdateUser(ctx context.Context, appID string, userID string, data map[string]interface{}) (*User, error) {
    payload := map[string]interface{}{
        "data": data,
    }

    req, err := http.NewRequestWithContext(ctx, "PUT", 
        fmt.Sprintf("%s/applications/%s/users/%s/data", c.BaseURL, appID, userID), 
        jsonReader(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doUserRequest(req)
}

func (c *Client) PatchUser(ctx context.Context, appID string, userID string, data map[string]interface{}) (*User, error) {
    payload := map[string]interface{}{
        "data": data,
    }

    req, err := http.NewRequestWithContext(ctx, "PATCH", 
        fmt.Sprintf("%s/applications/%s/users/%s/data", c.BaseURL, appID, userID), 
        jsonReader(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doUserRequest(req)
}

func (c *Client) GetUserField(ctx context.Context, appID string, userID string, field string) (interface{}, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", 
        fmt.Sprintf("%s/applications/%s/users/%s/data/fields/%s", c.BaseURL, appID, userID, field), 
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    resp, err := c.doRequest(req)
    if err != nil {
        return nil, err
    }

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return result["value"], nil
}

func (c *Client) UpdateUserField(ctx context.Context, appID string, userID string, field string, value interface{}) error {
    form := new(bytes.Buffer)
    writer := multipart.NewWriter(form)
    
    if err := writer.WriteField("value", fmt.Sprintf("%v", value)); err != nil {
        return NewError(ErrAPI, "failed to write form field", err)
    }
    writer.Close()

    req, err := http.NewRequestWithContext(ctx, "PUT", 
        fmt.Sprintf("%s/applications/%s/users/%s/data/fields/%s", c.BaseURL, appID, userID, field), 
        form)
    if err != nil {
        return NewError(ErrAPI, "failed to create request", err)
    }

    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    _, err = c.doRequest(req)
    return err
}

func jsonReader(v interface{}) io.Reader {
    data, _ := json.Marshal(v)
    return bytes.NewReader(data)
}

func (c *Client) doUserRequest(req *http.Request) (*User, error) {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &user, nil
}