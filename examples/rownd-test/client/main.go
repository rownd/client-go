package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    
    "github.com/rgthelen/rownd-go-test/pkg/rownd"
)

func main() {
    // Use environment variables or hardcoded token for testing
    token := os.Getenv("ROWND_TOKEN")
    if token == "" {
        token = "eyJhbGciOiJFZERTQSIsImtpZCI6InNpZy0xNjQ0OTM3MzYwIn0.eyJqdGkiOiI0NTVmOGExMS04NzdlLTRiMzctOTAyMC1lNjA0NGFiNTgyOTkiLCJhdWQiOlsiYXBwOmFwcF94a2J1bWw0OHFzM3R5eHhqanBheGVlbXYiXSwic3ViIjoidXNlcl9mMjhzd3QybzI5NmN5c3pkeXp0bmFyYjQiLCJpYXQiOjE3MzIwNzkwNDAsImh0dHBzOi8vYXV0aC5yb3duZC5pby9hcHBfdXNlcl9pZCI6InVzZXJfZjI4c3d0Mm8yOTZjeXN6ZHl6dG5hcmI0IiwiaHR0cHM6Ly9hdXRoLnJvd25kLmlvL2lzX3ZlcmlmaWVkX3VzZXIiOnRydWUsImh0dHBzOi8vYXV0aC5yb3duZC5pby9pc19hbm9ueW1vdXMiOnRydWUsImh0dHBzOi8vYXV0aC5yb3duZC5pby9hdXRoX2xldmVsIjoiZ3Vlc3QiLCJpc3MiOiJodHRwczovL2FwaS5yb3duZC5pbyIsImV4cCI6MTczMjA4MjY0MH0.fDN6OmmtYmI1I3BwjPvXvKkMq6CV0d7nY3_PtUN3XD39eEpdbjYnDAxXB0DiXXrf4l4AfxRjFR-H2mvaMmo-DQ"


    }

    // Create HTTP client with timeout
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    // Validate token
    validateReq, err := http.NewRequest("GET", "http://localhost:8080/validate", nil)
    if err != nil {
        log.Fatal(err)
    }
    validateReq.Header.Set("Authorization", "Bearer "+token)
    validateReq.Header.Set("Content-Type", "application/json")

    validateResp, err := client.Do(validateReq)
    if err != nil {
        log.Fatal(err)
    }
    defer validateResp.Body.Close()

    var validation rownd.TokenValidationResponse
    if err := json.NewDecoder(validateResp.Body).Decode(&validation); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Token validation: %+v\n", validation)

    // Get user data using validated token
    userReq, err := http.NewRequest("GET", "http://localhost:8080/user", nil)
    if err != nil {
        log.Fatal(err)
    }
    userReq.Header.Set("Authorization", "Bearer "+token)
    userReq.Header.Set("Content-Type", "application/json")

    userResp, err := client.Do(userReq)
    if err != nil {
        log.Fatal(err)
    }
    defer userResp.Body.Close()

    var user rownd.User
    if err := json.NewDecoder(userResp.Body).Decode(&user); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User data: %+v\n", user)
}