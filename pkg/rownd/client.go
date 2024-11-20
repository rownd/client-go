package rownd

import (
    "net/http"
    "time"
    "github.com/patrickmn/go-cache"
)

type Client struct {
    AppKey      string
    AppSecret   string
    BaseURL     string
    HTTPClient  *http.Client
    config      *ClientConfig
    cache       *cache.Cache
}

// NewClient creates a new Rownd client instance
func NewClient(config *ClientConfig) (*Client, error) {
    if err := validateConfig(config); err != nil {
        return nil, err
    }

    return &Client{
        AppKey:     config.AppKey,
        AppSecret:  config.AppSecret,
        BaseURL:    defaultString(config.BaseURL, "https://api.rownd.io"),
        HTTPClient: &http.Client{Timeout: defaultDuration(config.Timeout, 30*time.Second)},
        config:     config,
        cache:      cache.New(cache.NoExpiration, 10*time.Minute),
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