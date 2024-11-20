package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
    
    "github.com/rgthelen/rownd-go-test/pkg/rownd"
)

func main() {
    config := &rownd.ClientConfig{
        AppKey:    "YOUR API KEY",
        AppSecret: "YOUR APP SECRET",
        BaseURL:   "https://api.rownd.io",
    }
    
    client, err := rownd.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "No token provided", http.StatusUnauthorized)
            return
        }

        token = strings.TrimPrefix(token, "Bearer ")

        ctx := r.Context()
        validation, err := client.ValidateToken(ctx, token)
        if err != nil {
            log.Printf("Validation error: %v", err)
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(validation); err != nil {
            log.Printf("Error encoding response: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
    })

    http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "No token provided", http.StatusUnauthorized)
            return
        }

        token = strings.TrimPrefix(token, "Bearer ")

        ctx := r.Context()
        validation, err := client.ValidateToken(ctx, token)
        if err != nil {
            log.Printf("Validation error: %v", err)
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        user, err := client.GetUser(ctx, validation.UserID, validation)
        if err != nil {
            log.Printf("Error fetching user: %v", err)
            http.Error(w, "Error fetching user data", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(user); err != nil {
            log.Printf("Error encoding response: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }
    })

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}