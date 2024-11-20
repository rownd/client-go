package rownd

import (
    "net/http"
    "time"
    "github.com/patrickmn/go-cache"
    "fmt"
)

type Client struct {
    AppKey      string
    AppSecret   string
    AppID       string
    BaseURL     string
    HTTPClient  *http.Client
    cache       *cache.Cache
}

// NewClient creates a new Rownd client instance
func NewClient(config *ClientConfig) (*Client, error) {
    if config.AppKey == "" || config.AppSecret == "" {
        return nil, fmt.Errorf("app key and secret are required")
    }

    baseURL := config.BaseURL
    if baseURL == "" {
        baseURL = "https://api.rownd.io"
    }

    return &Client{
        AppKey:     config.AppKey,
        AppSecret:  config.AppSecret,
        AppID:      config.AppID,
        BaseURL:    baseURL,
        HTTPClient: &http.Client{
            Timeout: config.Timeout,
        },
        cache: cache.New(5*time.Minute, 10*time.Minute),
    }, nil
}

func validateConfig(config *ClientConfig) error {
    if config.AppKey == "" {
        return NewError(ErrValidation, "app key is required", nil)
    }
    if config.AppSecret == "" {
        return NewError(ErrValidation, "app secret is required", nil)
    }
    return nil
}

func defaultString(val, defaultVal string) string {
    if val == "" {
        return defaultVal
    }
    return val
}

func defaultDuration(val, defaultVal time.Duration) time.Duration {
    if val == 0 {
        return defaultVal
    }
    return val
} 