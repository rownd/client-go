# Rownd SDK for Go

Use this library to integrate Rownd into your Go application. The SDK provides convenient methods for user authentication, token validation, and user management.

## Installation

```bash
go get github.com/rgthelen/rownd-go-test
```

## Usage

Here's a basic usage example:

```go
package main
import (
"context"
"github.com/rgthelen/rownd-go-test/pkg/rownd"
)
func main() {
client, err := rownd.NewClient(&rownd.ClientConfig{
AppKey: "YOUR_APP_KEY",
AppSecret: "YOUR_APP_SECRET",
AppID: "YOUR_APP_ID", // Optional: Used as fallback if not in token
})
if err != nil {
panic(err)
}
ctx := context.Background()
// Validate a token
tokenInfo, err := client.ValidateToken(ctx, "your-token")
if err != nil {
panic(err)
}
// Get user information
userInfo, err := client.GetUser(ctx, tokenInfo.UserID, tokenInfo)
if err != nil {
panic(err)
}
}
```

## API Reference

The SDK provides the following main methods:

### Token Validation
- `ValidateToken(ctx context.Context, token string) (*TokenValidationResponse, error)`

### User Management
- `GetUser(ctx context.Context, userID string, tokenInfo *TokenValidationResponse) (*User, error)`
- `UpdateUser(ctx context.Context, userID string, data map[string]interface{}) (*User, error)`
- `DeleteUser(ctx context.Context, appID string, userID string) error`

### Smart Links
- `CreateSmartLink(ctx context.Context, opts *SmartLinkOptions) (*SmartLink, error)`

### Group Management
- `CreateGroupInvite(ctx context.Context, appID string, groupID string, req *CreateGroupInviteRequest) (*GroupInviteResponse, error)`
- `GetGroupInvite(ctx context.Context, appID string, groupID string, inviteID string) (*GroupInvite, error)`
- `UpdateGroupInvite(ctx context.Context, appID string, groupID string, inviteID string, req *CreateGroupInviteRequest) (*GroupInvite, error)`
- `DeleteGroupInvite(ctx context.Context, appID string, groupID string, inviteID string) error`
- `ListGroupInvites(ctx context.Context, appID string, groupID string) (*GroupInviteListResponse, error)`
- `CreateGroupMember(ctx context.Context, appID string, groupID string, req *CreateGroupMemberRequest) (*GroupMember, error)`
- `GetGroupMember(ctx context.Context, appID string, groupID string, memberID string) (*GroupMember, error)`
- `UpdateGroupMember(ctx context.Context, appID string, groupID string, memberID string, req *CreateGroupMemberRequest) (*GroupMember, error)`
- `DeleteGroupMember(ctx context.Context, appID string, groupID string, memberID string) error`
- `ListGroupMembers(ctx context.Context, appID string, groupID string) (*GroupMemberListResponse, error)`

For detailed documentation and examples, visit our [official documentation](https://docs.rownd.io).