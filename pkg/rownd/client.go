package rownd

import (
    "net/http"
    "time"
)

type Client struct {
    AppKey      string
    AppSecret   string
    BaseURL     string
    HTTPClient  *http.Client
}

// NewClient creates a new Rownd client instance
func NewClient(appKey, appSecret string) *Client {
    return &Client{
        AppKey:     appKey,
        AppSecret:  appSecret,
        BaseURL:    "https://api.rownd.io",
        HTTPClient: &http.Client{
            Timeout: time.Second * 30,
        },
    }
} 