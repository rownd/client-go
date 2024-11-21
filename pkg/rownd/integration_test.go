package rownd

import (
    "context"
    "encoding/json"
    "strings"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    rowndtesting "github.com/rgthelen/rownd-go-test/pkg/rownd/testing"
)

func TestRowndIntegration(t *testing.T) {
    // Get test configuration
    testConfig := rowndtesting.GetTestConfig()
    var validToken string // Will be set after magic link redemption

    client, err := NewClient(&ClientConfig{
        AppKey:    testConfig.AppKey,
        AppSecret: testConfig.AppSecret,
        AppID:     testConfig.AppID,
        BaseURL:   testConfig.BaseURL,
        Timeout:   10 * time.Second,
    })
    if err != nil {
        t.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()

    // Start with smart links tests to get our token
    t.Run("smart links", func(t *testing.T) {
        var smartLinkUserID string

        t.Run("create magic link", func(t *testing.T) {
            opts := &SmartLinkOptions{
                Purpose:          "auth",
                VerificationType: "email",
                Data: map[string]interface{}{
                    "email": "testlink@example.com",
                    "first_name": "Test",
                },
                RedirectURL: "https://example.com/redirect",
                Expiration: "30d",
            }

            link, err := client.CreateSmartLink(ctx, opts)
            assert.NoError(t, err)
            assert.NotNil(t, link)
            assert.NotEmpty(t, link.Link)
            
            // Extract link ID and redeem it
            parts := strings.Split(link.Link, "/")
            linkID := parts[len(parts)-1]
            
            magicLinkResp, err := client.RedeemMagicLink(ctx, linkID)
            assert.NoError(t, err)
            assert.NotNil(t, magicLinkResp)
            
            // Store the token and user ID for subsequent tests and cleanup
            validToken = magicLinkResp.AccessToken
            smartLinkUserID = magicLinkResp.AppUserID
            
            t.Logf("Created and redeemed magic link for user: %s", smartLinkUserID)
        })

        // Add cleanup at the test suite level to ensure it runs after all tests
        t.Cleanup(func() {
            if smartLinkUserID != "" {
                err := client.DeleteUser(ctx, testConfig.AppID, smartLinkUserID)
                if err != nil {
                    t.Logf("Failed to cleanup magic link user %s: %v", smartLinkUserID, err)
                } else {
                    t.Logf("Cleaned up magic link user: %s", smartLinkUserID)
                }
            }
        })
    })

    // Token validation tests using magic link token
    t.Run("token validation", func(t *testing.T) {
        if validToken == "" {
            t.Fatal("No valid token available from magic link")
        }

        t.Run("validate token", func(t *testing.T) {
            tokenInfo, err := client.ValidateToken(ctx, validToken)
            assert.NoError(t, err)
            assert.NotNil(t, tokenInfo)
            assert.NotEmpty(t, tokenInfo.UserID)
        })

        t.Run("extract token claims", func(t *testing.T) {
            tokenInfo, err := client.ValidateToken(ctx, validToken)
            assert.NoError(t, err)
            assert.NotNil(t, tokenInfo)
            
            // Check claims
            userID, ok := tokenInfo.DecodedToken[CLAIM_USER_ID].(string)
            assert.True(t, ok, "User ID claim not found or not a string")
            assert.NotEmpty(t, userID)
            t.Logf("User ID from token: %s", userID)
            
            isVerified, ok := tokenInfo.DecodedToken[CLAIM_IS_VERIFIED_USER].(bool)
            assert.True(t, ok, "is_verified_user claim not found or not a boolean")
            t.Logf("User verified status: %v", isVerified)
            
            authLevel, ok := tokenInfo.DecodedToken[CLAIM_AUTH_LEVEL].(string)
            assert.True(t, ok, "auth_level claim not found or not a string")
            assert.NotEmpty(t, authLevel)
            t.Logf("Auth level: %s", authLevel)
        })
    })

    // Test user management with auto-generated UUID
    var createdUserID string

    t.Run("create user with auto UUID", func(t *testing.T) {
        userData := map[string]interface{}{
            "data": map[string]interface{}{
                "email": "test@example.com",
                "first_name": "Test",
                "last_name": "User",
            },
        }

        // Log the test payload
        payloadBytes, _ := json.MarshalIndent(userData, "", "  ")
        t.Logf("Test Payload: %s", string(payloadBytes))

        // Don't provide a userID, let Rownd generate one
        user, err := client.UpdateUser(ctx, testConfig.AppID, "", userData)
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.NotEmpty(t, user.ID)
        assert.Equal(t, "test@example.com", user.Data["email"])
        
        createdUserID = user.ID
    })

    t.Run("edit user", func(t *testing.T) {
        updatedData := map[string]interface{}{
            "data": map[string]interface{}{
                "email": "test@example.com",
                "first_name": "Updated",
                "last_name": "Name",
            },
        }

        user, err := client.UpdateUser(ctx, testConfig.AppID, createdUserID, updatedData)
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, "Updated", user.Data["first_name"])
        assert.Equal(t, "Name", user.Data["last_name"])
        assert.Equal(t, "test@example.com", user.Data["email"])
    })

    t.Run("update single field", func(t *testing.T) {
        // Update just the first_name field
        err := client.UpdateUserField(ctx, testConfig.AppID, createdUserID, "first_name", "SingleField")
        assert.NoError(t, err)

        // Verify the change
        user, err := client.GetUser(ctx, createdUserID, nil)
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, "SingleField", user.Data["first_name"], "First name should be updated")
        assert.Equal(t, "Name", user.Data["last_name"], "Last name should be unchanged")
        assert.Equal(t, "test@example.com", user.Data["email"], "Email should be unchanged")
    })

    t.Run("get user", func(t *testing.T) {
        // First get token info
        tokenInfo, err := client.ValidateToken(ctx, validToken)
        assert.NoError(t, err)

        user, err := client.GetUser(ctx, createdUserID, tokenInfo)
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, createdUserID, user.ID)
    })


    t.Run("delete user", func(t *testing.T) {
        err := client.DeleteUser(ctx, testConfig.AppID, createdUserID)
        assert.NoError(t, err)

        // Verify user is deleted
        tokenInfo, err := client.ValidateToken(ctx, validToken)
        assert.NoError(t, err)
        
        _, err = client.GetUser(ctx, createdUserID, tokenInfo)
        assert.Error(t, err)
    })

    // Group management tests
    var groupID string
    var groupName = "Test Group"

    t.Run("create group", func(t *testing.T) {
        req := &CreateGroupRequest{
            Name:            groupName,
            AdmissionPolicy: "open",
        }

        group, err := client.CreateGroup(ctx, testConfig.AppID, req)
        assert.NoError(t, err)
        assert.NotNil(t, group)
        assert.Equal(t, groupName, group.Name)
        
        groupID = group.ID
    })

    t.Run("list groups", func(t *testing.T) {
        groups, err := client.ListGroups(ctx, testConfig.AppID)
        assert.NoError(t, err)
        assert.NotNil(t, groups)
        assert.Greater(t, groups.TotalResults, 0)
        
        // Find our created group
        found := false
        for _, g := range groups.Results {
            if g.ID == groupID {
                found = true
                assert.Equal(t, groupName, g.Name)
                break
            }
        }
        assert.True(t, found, "Created group not found in list")
    })

    // Create a new user for group testing
    var testUserID string
    t.Run("create user for group", func(t *testing.T) {
        userData := map[string]interface{}{
            "data": map[string]interface{}{
                "email": "grouptest@example.com",
                "first_name": "Group",
                "last_name": "Test",
            },
        }

        user, err := client.UpdateUser(ctx, testConfig.AppID, "", userData)
        assert.NoError(t, err)
        assert.NotNil(t, user)
        testUserID = user.ID
    })

    var memberID string
    t.Run("add user to group", func(t *testing.T) {
        req := &CreateGroupMemberRequest{
            UserID: testUserID,
            Roles:  []string{"member", "owner"},
            State:  "active",
        }

        member, err := client.CreateGroupMember(ctx, testConfig.AppID, groupID, req)
        assert.NoError(t, err)
        assert.NotNil(t, member)
        assert.Equal(t, testUserID, member.UserID)
        assert.Contains(t, member.Roles, "owner")
        memberID = member.ID
    })

    t.Run("update group member", func(t *testing.T) {
        req := &CreateGroupMemberRequest{
            UserID: testUserID,
            Roles:  []string{"member", "owner", "admin"},
            State:  "active",
        }

        member, err := client.UpdateGroupMember(ctx, testConfig.AppID, groupID, memberID, req)
        assert.NoError(t, err)
        assert.NotNil(t, member)
        assert.Contains(t, member.Roles, "admin")
    })

    t.Run("create group invite", func(t *testing.T) {
        req := &CreateGroupInviteRequest{
            Email:       "invite@example.com",
            Roles:       []string{"member"},
            RedirectURL: "https://example.com/accept",
        }

        invite, err := client.CreateGroupInvite(ctx, testConfig.AppID, groupID, req)
        assert.NoError(t, err)
        assert.NotNil(t, invite)
        assert.NotEmpty(t, invite.Link)
        assert.Equal(t, "invite@example.com", invite.Invitation.UserLookupValue)
    })

    // Store the second user's ID for cleanup
    var secondUserID string

    t.Run("add second user to group", func(t *testing.T) {
        // Create another user
        userData := map[string]interface{}{
            "data": map[string]interface{}{
                "email": "grouptest2@example.com",
                "last_name": "Test2",
            },
        }

        user, err := client.UpdateUser(ctx, testConfig.AppID, "", userData)
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.NotEmpty(t, user.ID)
        secondUserID = user.ID  // Store the ID for cleanup

        // Add them to the group as owner
        memberReq := &CreateGroupMemberRequest{
            UserID: user.ID,
            Roles:  []string{"member", "owner"},
            State:  "active",
        }

        member, err := client.CreateGroupMember(ctx, testConfig.AppID, groupID, memberReq)
        assert.NoError(t, err)
        assert.NotNil(t, member)
        assert.Equal(t, user.ID, member.UserID)
        assert.Contains(t, member.Roles, "owner")
    })

    t.Run("list group members", func(t *testing.T) {
        members, err := client.ListGroupMembers(ctx, testConfig.AppID, groupID)
        assert.NoError(t, err)
        assert.NotNil(t, members)
        assert.Equal(t, 3, members.TotalResults)

        // Verify active members (should be 2)
        activeMembers := 0
        for _, member := range members.Results {
            if member.State == "active" {
                activeMembers++
                assert.Contains(t, member.Roles, "owner")
            }
        }
        assert.Equal(t, 2, activeMembers, "Expected to find two active group members")

        // Verify invite_pending member (from invite)
        pendingMembers := 0
        for _, member := range members.Results {
            if member.State == "invite_pending" {
                pendingMembers++
                assert.Contains(t, member.Roles, "member")
                assert.Equal(t, "invite@example.com", member.Profile["email"])
            }
        }
        assert.Equal(t, 1, pendingMembers, "Expected to find one pending group member")
    })

    t.Run("remove user from group", func(t *testing.T) {
        err := client.DeleteGroupMember(ctx, testConfig.AppID, groupID, memberID)
        assert.NoError(t, err)
    })

    t.Run("delete group", func(t *testing.T) {
        err := client.DeleteGroup(ctx, testConfig.AppID, groupID)
        assert.NoError(t, err)
    })

    t.Run("delete test user", func(t *testing.T) {
        err := client.DeleteUser(ctx, testConfig.AppID, testUserID)
        assert.NoError(t, err)
    })

    t.Run("delete second test user", func(t *testing.T) {
        if secondUserID == "" {
            t.Fatal("Second user ID not found")
        }
        err := client.DeleteUser(ctx, testConfig.AppID, secondUserID)
        assert.NoError(t, err)
        t.Logf("Deleted second test user: %s", secondUserID)
    })
}