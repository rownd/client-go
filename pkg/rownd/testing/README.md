# Rownd Go SDK Integration Tests

## Overview
This test suite performs end-to-end testing of the Rownd Go SDK, including authentication, user management, and group operations.

## Prerequisites
- Go 1.19 or higher
- A Rownd account with API access
- `.env` file in the project root

## Environment Setup

1. Create a `.env` file in the project root with the following variables:
```env
ROWND_APP_KEY=your_app_key
ROWND_APP_SECRET=your_app_secret
ROWND_APP_ID=your_app_id
ROWND_BASE_URL=https://api.rownd.io
```

2. Install dependencies:
```bash
go mod download
```

## Running Tests

Run all integration tests:
```bash
go test -v ./pkg/rownd
```

Run a specific test:
```bash
go test -v -run TestRowndIntegration/smart_links ./pkg/rownd
```

## Test Structure

The integration tests cover:
1. Smart Links & Authentication
   - Creating magic links
   - Redeeming tokens
   - Token validation

2. User Management
   - Creating users
   - Updating user data
   - Deleting users

3. Group Management
   - Creating groups
   - Managing members
   - Group invitations

## Important Notes

- Tests run sequentially and depend on each other
- Each test cleans up its resources
- The smart link test creates a token used by subsequent tests
- Group tests create temporary users that are cleaned up afterward

## Troubleshooting

Common issues:

1. Environment Variables
   - Ensure all required variables are in `.env`
   - Check that app credentials are valid

2. API Access
   - Verify your IP is allowlisted
   - Confirm app key/secret have necessary permissions

3. Test Failures
   - Check API response in test logs
   - Verify token expiration times
   - Ensure cleanup from previous test runs

## Reference

Key files:
- Integration tests: `pkg/rownd/integration_test.go`
- Test utilities: `pkg/rownd/testing/test_utils.go`
- Auth types: `pkg/rownd/auth.go`

For more details on the test implementation, see:
```go:pkg/rownd/integration_test.go
startLine: 14
endLine: 375
```
