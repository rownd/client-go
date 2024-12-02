# Rownd SDK for Go

A comprehensive Go SDK for integrating Rownd authentication, user management, and group management into your applications.

## Features

- Token validation and management
- User authentication and management
- Group management with member and invite handling
- Smart link generation
- Full support for Rownd's REST API

## Installation

```bash
go get github.com/rgthelen/rownd-go-test
```

## Quick Start
```go
package main
import (
"context"
"log"
"github.com/rgthelen/rownd-go-test/pkg/rownd"
)
func main() {
// Initialize the client
client, err := rownd.NewClient(&rownd.ClientConfig{
AppKey: "YOUR_APP_KEY",
AppSecret: "YOUR_APP_SECRET",
AppID: "YOUR_APP_ID", // Optional: Used as fallback if not in token
})
if err != nil {
log.Fatal(err)
}
ctx := context.Background()
// Example: Create and manage a user
userData := map[string]interface{}{
"email": "user@example.com",
"first_name": "John",
"last_name": "Doe",
}
user, err := client.UpdateUser(ctx, "YOUR_APP_ID", "", userData)
if err != nil {
log.Fatal(err)
}
log.Printf("Created user: %v", user.ID)
}
```

## Usage Examples

### Token Validation
```go
tokenInfo, err := client.ValidateToken(ctx, "your-token")
if err != nil {
log.Fatal(err)
}
```

### User Management
```go
// Get user
user, err := client.GetUser(ctx, "user_id", tokenInfo)

// Update user
updatedUser, err := client.UpdateUser(ctx, appID, userID, userData)

// List/lookup users
users, err := client.LookupUsers(ctx, appID, &rownd.UserListOptions{
    LookupValue: "user@example.com",
    Fields: []string{"email"},
})

// Delete user
err := client.DeleteUser(ctx, appID, userID)
```

#### List User Options
- `LookupValue`: Value to search for (email/phone)
- `Fields`: Fields to include in response
- `PageSize`: Number of results per page (max 1000)
- `After`: ID of last resource from previous page
- `Sort`: Sort direction ("asc" or "desc")
- `IncludeDuplicates`: Include multiple matches

### Group Management
```go
// Create a group
group, err := client.CreateGroup(ctx, appID, &rownd.CreateGroupRequest{
Name: "Engineering Team",
AdmissionPolicy: "invite_only",
Meta: map[string]interface{}{
"department": "Engineering",
},
})
// Add member to group
member, err := client.CreateGroupMember(ctx, appID, groupID, &rownd.CreateGroupMemberRequest{
UserID: "user_123",
Roles: []string{"admin", "member"},
State: "active",
})
// Create group invite
invite, err := client.CreateGroupInvite(ctx, appID, groupID, &rownd.CreateGroupInviteRequest{
Email: "new@example.com",
Roles: []string{"member"},
RedirectURL: "/welcome",
})
```

#### Group Options
- `Name`: The name of the group.
- `AdmissionPolicy`: The policy for admitting new members (e.g., "invite_only", "open").
- `Meta`: Optional metadata associated with the group, such as department or other custom fields.

#### Important Notes

- **Group Management**: Groups must have at least one owner; ownership must be transferred before an owner can be deleted. A group must have at least one member (after a member is added). If you wish to delete the last member from a group, you must delete the group.

### Smart Links
Smart Links provide a way to create authentication and verification links for users. They can be used for magic link authentication, email verification, and group invitations.

```go
// Create a magic link for authentication
smartLink, err := client.CreateSmartLink(ctx, &rownd.SmartLinkOptions{
Purpose: "auth",
VerificationType: "email",
Data: map[string]interface{}{
"email": "user@example.com",
},
RedirectURL: "https://your-app.com/auth-callback",
})
if err != nil {
log.Fatal(err)
}
fmt.Printf("Magic link: %s\n", smartLink.Link)
// Create a group invitation link
smartLink, err := client.CreateSmartLink(ctx, &rownd.SmartLinkOptions{
Purpose: "auth",
VerificationType: "email",
Data: map[string]interface{}{
"email": "newmember@example.com",
},
RedirectURL: "https://your-app.com/groups",
GroupToJoin: "group_123",
Expiration: "7d", // Link expires in 7 days
})
```

#### Smart Link Options
- `Purpose`: The purpose of the link ("auth" for authentication)
- `VerificationType`: Type of verification ("email" or "phone")
- `Data`: Additional data to include (email, phone, etc.)
- `RedirectURL`: Where to redirect after link is used
- `UserID`: Optional user ID to associate with the link
- `Expiration`: Optional expiration time (e.g., "30d", "24h")
- `GroupToJoin`: Optional group ID for group invitations

## API Reference

### Token Validation
- `ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error)`

### User Management
- `GetUser(ctx context.Context, userID string, tokenInfo *TokenValidationResponse) (*User, error)`
- `UpdateUser(ctx context.Context, appID string, userID string, userData map[string]interface{}) (*User, error)`
- `DeleteUser(ctx context.Context, appID string, userID string) error`
- `GetUserField(ctx context.Context, appID string, userID string, field string) (interface{}, error)`
- `UpdateUserField(ctx context.Context, appID string, userID string, field string, value interface{}) error`

### Group Management
- `CreateGroup(ctx context.Context, appID string, req *CreateGroupRequest) (*Group, error)`
- `GetGroup(ctx context.Context, appID string, groupID string) (*Group, error)`
- `DeleteGroup(ctx context.Context, appID string, groupID string) error`
- `ListGroups(ctx context.Context, appID string) (*GroupListResponse, error)`

### Group Members
- `CreateGroupMember(ctx context.Context, appID string, groupID string, req *CreateGroupMemberRequest) (*GroupMember, error)`
- `GetGroupMember(ctx context.Context, appID string, groupID string, memberID string) (*GroupMember, error)`
- `UpdateGroupMember(ctx context.Context, appID string, groupID string, memberID string, req *CreateGroupMemberRequest) (*GroupMember, error)`
- `DeleteGroupMember(ctx context.Context, appID string, groupID string, memberID string) error`
- `ListGroupMembers(ctx context.Context, appID string, groupID string) (*GroupMemberListResponse, error)`

### Group Invites
- `CreateGroupInvite(ctx context.Context, appID string, groupID string, req *CreateGroupInviteRequest) (*GroupInviteResponse, error)`
- `GetGroupInvite(ctx context.Context, appID string, groupID string, inviteID string) (*GroupInvite, error)`
- `UpdateGroupInvite(ctx context.Context, appID string, groupID string, inviteID string, req *CreateGroupInviteRequest) (*GroupInvite, error)`
- `DeleteGroupInvite(ctx context.Context, appID string, groupID string, inviteID string) error`
- `ListGroupInvites(ctx context.Context, appID string, groupID string) (*GroupInviteListResponse, error)`

## Error Handling

The SDK provides detailed error information through the `Error` type. Example error handling:

```go
if err != nil {
switch rownd.GetErrorCode(err) {
case rownd.ErrValidation:
log.Printf("Validation error: %v", err)
case rownd.ErrAPI:
log.Printf("API error: %v", err)
case rownd.ErrNetwork:
log.Printf("Network error: %v", err)
default:
log.Printf("Unknown error: %v", err)
}
}
```
## Testing

Run the test suite:

```bash
go test ./...
```
## Documentation

For detailed documentation and examples, visit our [official documentation](https://docs.rownd.io).

## License

This project is licensed under the MIT License - see the LICENSE file for details.



