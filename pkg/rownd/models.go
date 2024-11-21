package rownd

import "time"

// User represents a Rownd user
type User struct {
    ID            string                 `json:"rownd_user"`
    State         string                 `json:"state"`
    AuthLevel     string                 `json:"auth_level"`
    Data          map[string]interface{} `json:"data"`
    VerifiedData  map[string]interface{} `json:"verified_data"`
    Groups        []GroupMembership      `json:"groups"`
    Meta          UserMeta               `json:"meta"`
    ConnectionMap map[string]interface{} `json:"connection_map"`
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
    AppID      string        // Optional: Used as fallback if not in token
    BaseURL    string
    Timeout    time.Duration
    RetryCount int
}

// UserUpdateRequest represents the request body for updating a user
type UserUpdateRequest struct {
    Data map[string]interface{} `json:"data"`
}

// WellKnownConfig represents the OAuth well-known configuration
type WellKnownConfig struct {
    Issuer                string   `json:"issuer"`
    AuthorizationEndpoint string   `json:"authorization_endpoint"`
    TokenEndpoint         string   `json:"token_endpoint"`
    JwksUri              string   `json:"jwks_uri"`
    ResponseTypesSupported []string `json:"response_types_supported"`
    SubjectTypesSupported []string `json:"subject_types_supported"`
    ScopesSupported      []string `json:"scopes_supported"`
}

type UserMeta struct {
    Created                      time.Time `json:"created"`
    Modified                     time.Time `json:"modified"`
    FirstSignIn                  time.Time `json:"first_sign_in"`
    FirstSignInMethod           string    `json:"first_sign_in_method"`
    LastSignIn                   time.Time `json:"last_sign_in"`
    LastSignInMethod            string    `json:"last_sign_in_method"`
    LastActive                   time.Time `json:"last_active"`
    LastPasskeyRegistrationPrompt time.Time `json:"last_passkey_registration_prompt"`
}

type GroupMembership struct {
    Group  Group       `json:"group"`
    Member GroupMember `json:"member"`
}