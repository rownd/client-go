package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/rgthelen/rownd-go-test/pkg/rownd"
)

func main() {
    client, err := rownd.NewClient(&rownd.ClientConfig{
        AppKey:    "your-app-key",
        AppSecret: "your-app-secret",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    
    // Validate a token
    validation, err := client.ValidateToken(ctx, "some-token")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Token validation: %+v\n", validation)

    // Get user
    user, err := client.GetUser("user-id")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User: %+v\n", user)

    // Update user
    userData := map[string]interface{}{
        "first_name": "John",
        "last_name": "Doe",
    }
    updatedUser, err := client.UpdateUser("user-id", userData)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Updated user: %+v\n", updatedUser)
}