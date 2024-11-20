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
    config      *ClientConfig
}

// NewClient creates a new Rownd client instance
func NewClient(config *ClientConfig) (*Client, error) {
    if err := validateConfig(config); err != nil {
        return nil, err
    }

    baseURL := config.BaseURL
    if baseURL == "" {
        baseURL = "https://api.rownd.io"
    }

    timeout := config.Timeout
    if timeout == 0 {
        timeout = 30 * time.Second
    }

    return &Client{
        AppKey:     config.AppKey,
        AppSecret:  config.AppSecret,
        BaseURL:    baseURL,
        HTTPClient: &http.Client{
            Timeout: timeout,
        },
        config: config,
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