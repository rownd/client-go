package rownd

type AuthInitRequest struct {
    Email             string                 `json:"email"`
    ContinueWithEmail bool                   `json:"continue_with_email"`
    Fingerprint       map[string]interface{} `json:"fingerprint,omitempty"`
    ReturnURL         string                 `json:"return_url"`
}

type AuthInitResponse struct {
    ChallengeID    string `json:"challenge_id"`
    ChallengeToken string `json:"challenge_token"`
    AuthTokens     *struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
    } `json:"auth_tokens,omitempty"`
}

type AuthCompleteRequest struct {
    Token       string `json:"token"`
    ChallengeID string `json:"challenge_id"`
    Email       string `json:"email"`
}

type AuthCompleteResponse struct {
    RedirectURL string `json:"redirect_url"`
}

type AuthTokens struct {
    AccessToken  string
    RefreshToken string
}