package rownd_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rownd/client-go/internal/testutils"
	"github.com/rownd/client-go/pkg/rownd"
	"github.com/stretchr/testify/assert"
)

func TestRowndUserOperations(t *testing.T) {
	var createdUser *rownd.User

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

	t.Run("create user with auto UUID", func(t *testing.T) {
		userData := map[string]interface{}{
			"email":      "test@example.com",
			"first_name": "Test",
			"last_name":  "User",
		}

		req := rownd.CreateOrUpdateUserRequest{
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
			Fields:       []string{"email", "first_name", "last_name", "user_id"}, // Request all needed fields
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
			UserID: createdUserID,
		})
		assert.NoError(t, err)
	})

	t.Run("user attributes operations", func(t *testing.T) {
		// Create a user for attribute testing
		userData := map[string]interface{}{
			"email":      fmt.Sprintf("test.attributes.%d@example.com", time.Now().UnixNano()),
			"first_name": "Attribute",
			"last_name":  "Test",
		}

		user, err := client.Users.CreateOrUpdate(ctx, rownd.CreateOrUpdateUserRequest{
			UserID: "__UUID__",
			Data:   userData,
			Attributes: map[string][]string{
				"rownd:app_variants": {"variant_1", "variant_2"},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, user.ID)
		testUserID := user.ID

		// Test GetAttributes method
		attrs := user.GetAttributes()
		assert.NotNil(t, attrs)
		assert.Equal(t, []string{"variant_1", "variant_2"}, attrs["rownd:app_variants"])

		// Test AddAttributes
		t.Run("add attributes", func(t *testing.T) {
			user, err := client.Users.AddAttributes(ctx, rownd.AddAttributesRequest{
				UserID: testUserID,
				Attributes: map[string][]string{
					"myapp:subscription_status": {"active"},
					"myapp:loyalty_points":      {"100"},
				},
			})
			assert.NoError(t, err)
			assert.NotNil(t, user)

			// Verify attributes were added
			attrs := user.GetAttributes()
			assert.Equal(t, []string{"variant_1", "variant_2"}, attrs["rownd:app_variants"])
			assert.Equal(t, []string{"active"}, attrs["myapp:subscription_status"])
			assert.Equal(t, []string{"100"}, attrs["myapp:loyalty_points"])
		})

		// Test updating existing attributes
		t.Run("update existing attributes", func(t *testing.T) {
			user, err := client.Users.AddAttributes(ctx, rownd.AddAttributesRequest{
				UserID: testUserID,
				Attributes: map[string][]string{
					"myapp:loyalty_points": {"200"},  // Update existing
					"myapp:tier":           {"gold"}, // Add new
				},
			})
			assert.NoError(t, err)
			assert.NotNil(t, user)

			attrs := user.GetAttributes()
			assert.Equal(t, []string{"200"}, attrs["myapp:loyalty_points"])
			assert.Equal(t, []string{"gold"}, attrs["myapp:tier"])
		})

		// Test DeleteAttributes
		t.Run("delete attributes", func(t *testing.T) {
			user, err := client.Users.DeleteAttributes(ctx, rownd.DeleteAttributesRequest{
				UserID:        testUserID,
				AttributeKeys: []string{"myapp:subscription_status", "myapp:tier"},
			})
			assert.NoError(t, err)
			assert.NotNil(t, user)

			// Verify attributes were deleted
			attrs := user.GetAttributes()
			assert.Equal(t, []string{"variant_1", "variant_2"}, attrs["rownd:app_variants"])
			assert.Equal(t, []string{"200"}, attrs["myapp:loyalty_points"])
			_, hasSubscription := attrs["myapp:subscription_status"]
			assert.False(t, hasSubscription)
			_, hasTier := attrs["myapp:tier"]
			assert.False(t, hasTier)
		})

		// Test validation errors
		t.Run("validation errors", func(t *testing.T) {
			// Test AddAttributes without UserID
			_, err := client.Users.AddAttributes(ctx, rownd.AddAttributesRequest{
				Attributes: map[string][]string{"test": {"value"}},
			})
			assert.Error(t, err)

			// Test AddAttributes without attributes
			_, err = client.Users.AddAttributes(ctx, rownd.AddAttributesRequest{
				UserID: testUserID,
			})
			assert.Error(t, err)

			// Test DeleteAttributes without UserID
			_, err = client.Users.DeleteAttributes(ctx, rownd.DeleteAttributesRequest{
				AttributeKeys: []string{"test"},
			})
			assert.Error(t, err)

			// Test DeleteAttributes without keys
			_, err = client.Users.DeleteAttributes(ctx, rownd.DeleteAttributesRequest{
				UserID: testUserID,
			})
			assert.Error(t, err)
		})

		// Cleanup
		err = client.Users.Delete(ctx, rownd.DeleteUserRequest{
			UserID: testUserID,
		})
		assert.NoError(t, err)
	})
}
