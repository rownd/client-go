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
"fmt"
"github.com/rgthelen/rownd-go-test/pkg/rownd"
)
func main() {
client := rownd.NewClient(&rownd.Config{
AppKey: "YOUR_ROWND_APP_KEY",
AppSecret: "YOUR_ROWND_APP_SECRET",
})
// Validate a token
tokenInfo, err := client.ValidateToken("your-token-here")
if err != nil {
panic(err)
}
// Get user information
userInfo, err := client.GetUser(tokenInfo.UserID)
if err != nil {
panic(err)
}
fmt.Printf("User data: %+v\n", userInfo.Data)
}
```

This example demonstrates how to initialize a client, validate a token, and retrieve user information. You can expand this code to include other Rownd features as needed.


## API Reference

The SDK provides the following main methods:

### Token Validation
- `ValidateToken(token string) (*TokenValidation, error)`

### User Management
- `GetUser(userID string) (*User, error)`
- `UpdateUser(userID string, data map[string]interface{}) (*User, error)`
- `DeleteUser(userID string) error`

### Smart Links
- `CreateSmartLink(opts *SmartLinkOptions) (*SmartLink, error)`

For detailed documentation and examples, visit our [official documentation](https://docs.rownd.io).



