package types

import (
    "context"
)

type ClientConfig struct {
    AppKey    string
    AppSecret string
    AppID     string
    BaseURL   string
}

type AuthInitRequest struct {
    Email             string
    ContinueWithEmail bool
    ReturnURL         string
}

type AuthInitResponse struct {
    ChallengeID    string      `json:"challenge_id"`
    ChallengeToken string      `json:"challenge_token"`
    AuthTokens     *AuthTokens `json:"auth_tokens,omitempty"`
}

type AuthTokens struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

type TokenValidationResponse struct {
    UserID      string `json:"user_id"`
    AccessToken string `json:"access_token"`
}

// Client interface defines the methods that must be implemented
type Client interface {
    InitiateAuth(ctx context.Context, req *AuthInitRequest) (*AuthInitResponse, error)
    ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error)
}