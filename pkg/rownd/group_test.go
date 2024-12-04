package rownd_test

import (
	"context"
	"testing"

	"github.com/rownd/client-go/internal/testutils"
	"github.com/rownd/client-go/pkg/rownd"
	"github.com/stretchr/testify/assert"
)

func TestRowndGroups(t *testing.T) {
	// Get test configuration
	testConfig := testutils.GetTestConfig()

	client, err := rownd.NewClient(
		rownd.WithAppKey(testConfig.AppKey),
		rownd.WithAppSecret(testConfig.AppSecret),
		rownd.WithBaseURL(testConfig.BaseURL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Group management tests
	var groupID string
	var groupName = "Test Group"

	t.Run("create group", func(t *testing.T) {
		group, err := client.Groups.Create(ctx, rownd.CreateGroupRequest{
			Name:            groupName,
			AdmissionPolicy: "invite_only",
		})
		println(err)
		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, groupName, group.Name)

		groupID = group.ID
	})

	t.Run("list groups", func(t *testing.T) {
		groups, err := client.Groups.List(ctx, rownd.ListGroupsRequest{})
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
			"email":      "grouptest@example.com",
			"first_name": "Group",
			"last_name":  "Test",
		}

		user, err := client.Users.CreateOrUpdate(ctx, rownd.CreateOrUpdateUserRequest{
			UserID: "__UUID__",
			Data:   userData,
		})
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		assert.NotNil(t, user)

		// Get the user ID from the data field
		testUserID = user.Data["user_id"].(string)
		t.Logf("Created test user with ID: %s", testUserID)

		// Verify the user ID
		if testUserID == "" {
			t.Fatal("User ID is empty")
		}
	})

	var memberID string
	t.Run("add user to group", func(t *testing.T) {
		t.Logf("Attempting to add user ID: %s to group ID: %s", testUserID, groupID)

		memberRequest := rownd.CreateGroupMemberRequest{
			GroupID: groupID,
			UserID:  testUserID,
			Roles:   []string{"member"},
			State:   "active",
		}

		member, err := client.GroupMembers.Create(ctx, memberRequest)
		if err != nil {
			t.Fatalf("Failed to create group member: %v", err)
		}

		assert.NotNil(t, member)
		assert.Equal(t, testUserID, member.UserID)
		memberID = member.ID
		t.Logf("Created member with ID: %s", memberID)
	})

	t.Run("update group member", func(t *testing.T) {
		assert.NotEmpty(t, memberID, "Member ID should not be empty")
		t.Logf("Updating member ID: %s", memberID)

		member, err := client.GroupMembers.Update(ctx, rownd.UpdateGroupMemberRequest{
			GroupID:  groupID,
			MemberID: memberID,
			UserID:   testUserID,
			Roles:    []string{"member", "owner", "admin"},
			State:    "active",
		})
		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Contains(t, member.Roles, "admin")
	})

	t.Run("create group invite", func(t *testing.T) {
		invite, err := client.GroupInvites.Create(ctx, rownd.CreateGroupInviteRequest{
			GroupID:     groupID,
			Email:       "invite@example.com",
			Roles:       []string{"member"},
			RedirectURL: "https://example.com/accept",
		})
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
			"email":     "grouptest2@example.com",
			"last_name": "Test2",
		}

		user, err := client.Users.CreateOrUpdate(ctx, rownd.CreateOrUpdateUserRequest{
			UserID: "__UUID__",
			Data:   userData,
		})
		assert.NoError(t, err)
		assert.NotNil(t, user)

		// Get the user ID from the data field
		secondUserID = user.Data["user_id"].(string)
		assert.NotEmpty(t, secondUserID)
		t.Logf("Created second user with ID: %s", secondUserID)

		// Add them to the group as owner
		member, err := client.GroupMembers.Create(ctx, rownd.CreateGroupMemberRequest{
			GroupID: groupID,
			UserID:  secondUserID,
			Roles:   []string{"member", "owner"},
			State:   "active",
		})
		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, secondUserID, member.UserID)
		assert.Contains(t, member.Roles, "owner")
	})

	t.Run("list group members", func(t *testing.T) {
		members, err := client.GroupMembers.List(ctx, rownd.ListGroupMembersRequest{
			GroupID: groupID,
		})
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
		t.Logf("Attempting to remove member ID: %s from group ID: %s", memberID, groupID)

		// List members before deletion
		beforeMembers, err := client.GroupMembers.List(ctx, rownd.ListGroupMembersRequest{
			GroupID: groupID,
		})
		assert.NoError(t, err)
		t.Logf("Members before deletion: %+v", beforeMembers)

		// Attempt deletion
		err = client.GroupMembers.Delete(ctx, rownd.DeleteGroupMemberRequest{
			GroupID:  groupID,
			MemberID: memberID,
		})
		assert.NoError(t, err)

		// List members after deletion to verify
		afterMembers, err := client.GroupMembers.List(ctx, rownd.ListGroupMembersRequest{
			GroupID: groupID,
		})
		assert.NoError(t, err)
		t.Logf("Members after deletion: %+v", afterMembers)

		// Verify member was removed
		for _, member := range afterMembers.Results {
			assert.NotEqual(t, memberID, member.ID, "Deleted member should not be present")
		}
	})

	t.Run("delete group", func(t *testing.T) {
		err := client.Groups.Delete(ctx, rownd.DeleteGroupRequest{
			GroupID: groupID,
		})
		assert.NoError(t, err)
	})

	t.Run("delete test user", func(t *testing.T) {
		err := client.Users.Delete(ctx, rownd.DeleteUserRequest{
			UserID: testUserID,
		})
		assert.NoError(t, err)
	})

	t.Run("delete second test user", func(t *testing.T) {
		if secondUserID == "" {
			t.Fatal("Second user ID not found")
		}
		err := client.Users.Delete(ctx, rownd.DeleteUserRequest{
			UserID: secondUserID,
		})
		assert.NoError(t, err)
		t.Logf("Deleted second test user: %s", secondUserID)
	})
}
