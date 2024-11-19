package rownd

import "time"

// User represents a Rownd user
type User struct {
    ID        string                 `json:"id"`
    Data      map[string]interface{} `json:"data"`
    CreatedAt time.Time             `json:"created_at"`
    UpdatedAt time.Time             `json:"updated_at"`
}

// TokenValidation represents the response from token validation
type TokenValidation struct {
    Valid    bool      `json:"valid"`
    UserID   string    `json:"user_id"`
    AppID    string    `json:"app_id"`
    IssuedAt time.Time `json:"iat"`
    ExpireAt time.Time `json:"exp"`
}

// APIResponse represents a generic API response
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// ClientConfig represents the configuration for the Rownd client
type ClientConfig struct {
    AppKey     string
    AppSecret  string
    BaseURL    string
    Timeout    time.Duration
    RetryCount int
}

// UserUpdateRequest represents the request body for updating a user
type UserUpdateRequest struct {
    Data map[string]interface{} `json:"data"`
}