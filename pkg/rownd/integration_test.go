package rownd_test

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/rgthelen/rownd-go-test/pkg/rownd"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupTestServer() *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/hub/auth/init":
            json.NewEncoder(w).Encode(map[string]interface{}{
                "challenge_id": "test-challenge-123",
                "challenge_token": "test-challenge-token",
            })
        case "/hub/auth/complete":
            json.NewEncoder(w).Encode(map[string]interface{}{
                "redirect_url": "https://app.rownd.io/callback#access_token=test-access-token&refresh_token=test-refresh-token",
            })
        case "/applications/test-app-id/users/test-user-1":
            json.NewEncoder(w).Encode(&rownd.User{
                ID: "test-user-1",
                Data: map[string]interface{}{
                    "email": "test@example.com",
                },
            })
        }
    }))
}

func TestAuthFlow(t *testing.T) {
    server := setupTestServer()
    defer server.Close()

    client, err := rownd.NewClient(&rownd.ClientConfig{
        AppKey:    "test-app-key",
        AppSecret: "test-app-secret",
        AppID:     "test-app-id",
        BaseURL:   server.URL,
    })
    require.NoError(t, err)

    ctx := context.Background()

    // Initialize authentication
    initResp, err := client.InitiateAuth(ctx, &rownd.AuthInitRequest{
        Email:             "test@example.com",
        ContinueWithEmail: true,
        ReturnURL:        "https://localhost:8787/static/test",
    })
    require.NoError(t, err)
    assert.NotEmpty(t, initResp.ChallengeID)
    assert.NotEmpty(t, initResp.ChallengeToken)

    // Complete authentication
    completeResp, err := client.CompleteAuth(ctx, &rownd.AuthCompleteRequest{
        Token:       initResp.ChallengeToken,
        ChallengeID: initResp.ChallengeID,
        Email:       "test@example.com",
    })
    require.NoError(t, err)
    assert.Contains(t, completeResp.RedirectURL, "access_token")

    // Extract tokens from redirect URL
    tokens, err := rownd.ParseAuthRedirect(completeResp.RedirectURL)
    require.NoError(t, err)
    assert.NotEmpty(t, tokens.AccessToken)
    assert.NotEmpty(t, tokens.RefreshToken)

    // Validate the token
    tokenInfo, err := client.ValidateToken(ctx, tokens.AccessToken)
    require.NoError(t, err)
    assert.NotEmpty(t, tokenInfo.UserID)

    // Test user operations with the token
    t.Run("user operations", func(t *testing.T) {
        // Get user
        user, err := client.GetUser(ctx, tokenInfo.UserID)
        require.NoError(t, err)
        assert.Equal(t, "test@example.com", user.Data["email"])

        // Update user
        updatedUser, err := client.UpdateUser(ctx, tokenInfo.UserID, map[string]interface{}{
            "first_name": "Test",
            "last_name":  "User",
        })
        require.NoError(t, err)
        assert.Equal(t, "Test", updatedUser.Data["first_name"])
    })
}