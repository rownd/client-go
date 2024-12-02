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
    Email             string `json:"email"`
    ContinueWithEmail bool   `json:"continue_with_email"`
    ReturnURL         string `json:"return_url"`
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

type SmartLinkOptions struct {
    Purpose          string                 `json:"purpose"`
    VerificationType string                 `json:"verification_type"`
    Data            map[string]interface{} `json:"data"`
    RedirectURL     string                 `json:"redirect_url"`
    Expiration      string                 `json:"expiration"`
}

type CreateGroupRequest struct {
    Name            string `json:"name"`
    AdmissionPolicy string `json:"admission_policy"`
}

type CreateGroupMemberRequest struct {
    UserID string   `json:"user_id"`
    Roles  []string `json:"roles"`
    State  string   `json:"state"`
}

type CreateGroupInviteRequest struct {
    Email       string   `json:"email"`
    Roles       []string `json:"roles"`
    RedirectURL string   `json:"redirect_url"`
}

type AuthCompleteRequest struct {
    ChallengeID    string `json:"challenge_id"`
    ChallengeToken string `json:"challenge_token"`
}

// Client interface defines the methods that must be implemented
type Client interface {
    InitiateAuth(ctx context.Context, req *AuthInitRequest) (*AuthInitResponse, error)
    ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error)
}