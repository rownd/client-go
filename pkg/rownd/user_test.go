package rownd_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rgthelen/rownd-go-test/internal/testutils"
	"github.com/rgthelen/rownd-go-test/pkg/rownd"
	"github.com/stretchr/testify/assert"
)


func TestRowndUserOperations(t *testing.T) {
	var createdUser *rownd.User

	// Get test configuration
	testConfig := testutils.GetTestConfig()

	client, err := rownd.NewClient(
		rownd.WithAppKey(testConfig.AppKey),
		rownd.WithAppSecret(testConfig.AppSecret),
		rownd.WithAppID(testConfig.AppID),
		rownd.WithBaseURL(testConfig.BaseURL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("create user with auto UUID", func(t *testing.T) {
		userData := map[string]interface{}{
			"email":      "test@example.com",
			"first_name": "Test",
			"last_name":  "User",
		}

		req := rownd.CreateOrUpdateUserRequest{
			AppID:  testConfig.AppID,
			UserID: "__UUID__",
			Data:   userData,
		}

		user, err := client.Users.CreateOrUpdate(ctx, req)
		if err != nil {
			t.Fatal(err)
		}

		createdUser = user
		t.Logf("Created user with ID: %s", user.ID)

		if user.ID == "__UUID__" || user.ID == "" {
			t.Fatal("Failed to get valid user ID from creation")
		}
	})

	t.Run("edit user", func(t *testing.T) {
		if createdUser == nil {
			t.Fatal("No user available to update")
		}

		updatedData := map[string]interface{}{
			"email":      "test@example.com",
			"first_name": "Updated",
			"last_name":  "Name",
		}

		user, err := client.Users.CreateOrUpdate(ctx, rownd.CreateOrUpdateUserRequest{
			AppID:  testConfig.AppID,
			UserID: createdUser.ID,
			Data:   updatedData,
		})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Updated", user.Data["first_name"])
	})

	t.Run("update_single_field", func(t *testing.T) {
		if createdUser == nil {
			t.Fatal("No user available to update")
		}

		t.Logf("Updating user with ID: %s", createdUser.ID)
		err := client.UserFields.Update(ctx, rownd.UpdateUserFieldRequest{
			AppID:  testConfig.AppID,
			UserID: createdUser.ID,
			Field:  "first_name",
			Value:  "SingleField",
		})
		assert.NoError(t, err)
	})

	t.Run("get user", func(t *testing.T) {
		if createdUser == nil {
			t.Fatal("No user available to get")
		}

		user, err := client.Users.Get(ctx, rownd.GetUserRequest{
			AppID:  testConfig.AppID,
			UserID: createdUser.ID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
	})

	t.Run("delete user", func(t *testing.T) {
		if createdUser == nil {
			t.Fatal("No user available to delete")
		}

		t.Logf("Attempting to delete user ID: %s", createdUser.ID)
		err := client.Users.Delete(ctx, rownd.DeleteUserRequest{
			AppID:  testConfig.AppID,
			UserID: createdUser.ID,
		})
		assert.NoError(t, err)
	})

	t.Run("lookup user", func(t *testing.T) {
		// Create a random email for testing
		randomEmail := fmt.Sprintf("test.lookup.%d@example.com", time.Now().UnixNano())

		// Create a new user with the random email
		userData := map[string]interface{}{
			"email":      randomEmail,
			"first_name": "Lookup",
			"last_name":  "Test",
		}

		user, err := client.Users.CreateOrUpdate(ctx, rownd.CreateOrUpdateUserRequest{
			AppID:  testConfig.AppID,
			UserID: "__UUID__",
			Data:   userData,
		})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, user.ID)
		createdUserID := user.ID

		// Add a small delay to allow for data propagation
		time.Sleep(2 * time.Second)

		// Lookup the user by email with all fields
		users, err := client.Users.List(ctx, rownd.ListUsersRequest{
			AppID:        testConfig.AppID,
			Fields:       []string{"email", "first_name", "last_name", "user_id"},  // Request all needed fields
			LookupFilter: []string{randomEmail},
		})
		assert.NoError(t, err)
		assert.NotNil(t, users)

		// Debug output
		t.Logf("Lookup results - Total: %d, Results length: %d", users.TotalResults, len(users.Results))
		t.Logf("Looking up email: %s", randomEmail)
		if len(users.Results) > 0 {
			t.Logf("First result raw: %+v", users.Results[0])
			t.Logf("First result ID: %s", users.Results[0].GetID())
			t.Logf("First result Data: %+v", users.Results[0].Data)
		}

		// Check if we found any users
		if assert.Greater(t, len(users.Results), 0, "No users found for email %s", randomEmail) {
			// Verify the looked-up user matches the created user
			foundUser := users.Results[0]
			assert.Equal(t, createdUserID, foundUser.ID)
			assert.Equal(t, randomEmail, foundUser.Data["email"])
			assert.Equal(t, "Lookup", foundUser.Data["first_name"])
			assert.Equal(t, "Test", foundUser.Data["last_name"])
		}

		// Add delay before cleanup
		time.Sleep(5 * time.Second)

		// Cleanup
		err = client.Users.Delete(ctx, rownd.DeleteUserRequest{
			AppID:  testConfig.AppID,
			UserID: createdUserID,
		})
		assert.NoError(t, err)
	})
}
