package main

import (
    "fmt"
    "log"
    
    "github.com/your-username/rownd-go"
)

func main() {
    client := rownd.NewClient("your-app-key", "your-app-secret")

    // Validate a token
    validation, err := client.ValidateToken("some-token")
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