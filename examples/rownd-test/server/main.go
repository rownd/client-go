package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "context"
	"os"
    
    "github.com/rgthelen/rownd-go-test/pkg/rownd"
)

// Global variables for handlers to use
var (
    client *rownd.Client
    appID  string
)

type ErrorResponse struct {
    Status  int    `json:"status"`
    Message string `json:"message"`
    Detail  string `json:"detail,omitempty"`
}

func writeError(w http.ResponseWriter, status int, message string, detail string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{
        Status:  status,
        Message: message,
        Detail:  detail,
    })
}

func main() {
    // Set global app ID
    appID = "ROWND_APP_ID"

    config := &rownd.ClientConfig{
        AppKey:    "ROWND_API_KEY",
        AppSecret: "ROWND_API_SECRET",
        BaseURL:   "https://api.rownd.io",
    }
    
    var err error
    client, err = rownd.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }

    // Serve static files
    fs := http.FileServer(http.Dir("client/static"))
    http.Handle("/", fs)

    // Add CORS headers middleware
    corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next(w, r)
        }
    }

    // Add JWT validation middleware
    authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                writeError(w, http.StatusUnauthorized, "Authentication required", "No token provided")
                return
            }
            token = strings.TrimPrefix(token, "Bearer ")
            
            validation, err := client.ValidateToken(r.Context(), token)
            if err != nil {
                log.Printf("Validation error: %v", err)
                writeError(w, http.StatusUnauthorized, "Invalid token", err.Error())
                return
            }
            
            ctx := context.WithValue(r.Context(), "validation", validation)
            next.ServeHTTP(w, r.WithContext(ctx))
        }
    }

    // Existing endpoints
    http.HandleFunc("/validate", corsMiddleware(authMiddleware(validateHandler)))
    http.HandleFunc("/user", corsMiddleware(authMiddleware(userHandler)))

    // New group endpoints
    http.HandleFunc("/groups", corsMiddleware(authMiddleware(groupsHandler)))
    http.HandleFunc("/groups/create", corsMiddleware(authMiddleware(createGroupHandler)))
    http.HandleFunc("/groups/invite", corsMiddleware(authMiddleware(createGroupInviteHandler)))

    // New user field endpoints
    http.HandleFunc("/user/field/", corsMiddleware(authMiddleware(updateUserFieldHandler)))

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    validation := ctx.Value("validation").(*rownd.TokenValidationResponse)

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(validation); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    validation := ctx.Value("validation").(*rownd.TokenValidationResponse)
    
    user, err := client.GetUser(ctx, validation.UserID, validation)
    if err != nil {
        log.Printf("Error fetching user: %v", err)
        writeError(w, http.StatusInternalServerError, "Failed to fetch user data", err.Error())
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    groups, err := client.ListGroups(ctx, appID)
    if err != nil {
        log.Printf("Error listing groups: %v", err)
        http.Error(w, "Failed to list groups", http.StatusInternalServerError)
        return
    }

    type GroupWithMembers struct {
        rownd.Group
        Members []rownd.GroupMember `json:"members"`
    }

    type GroupsResponse struct {
        TotalResults int              `json:"total_results"`
        Results      []GroupWithMembers `json:"results"`
    }

    // Get members for each group
    groupsWithMembers := []GroupWithMembers{}
    for _, group := range groups.Results {
        members, err := client.ListGroupMembers(ctx, appID, group.ID)
        if err != nil {
            log.Printf("Error listing group members for group %s: %v", group.ID, err)
            continue
        }

        groupsWithMembers = append(groupsWithMembers, GroupWithMembers{
            Group:   group,
            Members: members.Results,
        })
    }

    response := GroupsResponse{
        TotalResults: len(groupsWithMembers),
        Results:      groupsWithMembers,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func createGroupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    validation := r.Context().Value("validation").(*rownd.TokenValidationResponse)

    var req rownd.CreateGroupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate required fields
    if req.Name == "" {
        http.Error(w, "Group name is required", http.StatusBadRequest)
        return
    }

    // Set default admission policy if not provided
    if req.AdmissionPolicy == "" {
        req.AdmissionPolicy = "open"
    }

    // Add metadata about creator
    if req.Meta == nil {
        req.Meta = make(map[string]interface{})
    }
    req.Meta["created_by_user_id"] = validation.UserID

    ctx := r.Context()
    group, err := client.CreateGroup(ctx, appID, &req)
    if err != nil {
        log.Printf("Error creating group: %v", err)
        http.Error(w, "Failed to create group", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(group)
}

func createGroupInviteHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    validation := r.Context().Value("validation").(*rownd.TokenValidationResponse)

    groupID := r.URL.Query().Get("group_id")
    if groupID == "" {
        http.Error(w, "group_id is required", http.StatusBadRequest)
        return
    }

    var req rownd.CreateGroupInviteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // If no body provided, use defaults
        req = rownd.CreateGroupInviteRequest{
            UserID: validation.UserID,
            Roles: []string{"member"},
            RedirectURL: "http://localhost:8080",
        }
    } else {
        // Ensure required fields
        if req.Roles == nil {
            req.Roles = []string{"member"}
        }
        if req.RedirectURL == "" {
            req.RedirectURL = "http://localhost:8080"
        }
        if req.UserID == "" {
            req.UserID = validation.UserID
        }
    }

    ctx := r.Context()
    invite, err := client.CreateGroupInvite(ctx, appID, groupID, &req)
    if err != nil {
        log.Printf("Error creating group invite: %v", err)
        http.Error(w, "Failed to create group invite", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(invite)
}

func updateUserFieldHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "PUT" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ctx := r.Context()
    validation := ctx.Value("validation").(*rownd.TokenValidationResponse)
    
    // Get the field name from the URL path
    field := strings.TrimPrefix(r.URL.Path, "/user/field/")
    
    var req struct {
        Value interface{} `json:"value"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Get current user data
    user, err := client.GetUser(ctx, validation.UserID, validation)
    if err != nil {
        log.Printf("Error fetching user: %v", err)
        http.Error(w, "Error fetching user data", http.StatusInternalServerError)
        return
    }

    // Create data map with the new field value
    data := map[string]interface{}{
        field: req.Value,
    }

    // Add either email or anonymous_id for verification
    if email, ok := user.Data["email"].(string); ok && email != "" {
        data["email"] = email
    } else if anonID, ok := user.Data["anonymous_id"].(string); ok && anonID != "" {
        data["anonymous_id"] = anonID
    } else {
        http.Error(w, "No email or anonymous_id found for user", http.StatusBadRequest)
        return
    }

    // Use PatchUser to update the fields
    updatedUser, err := client.PatchUser(ctx, appID, validation.UserID, data)
    if err != nil {
        log.Printf("Error updating user field: %v", err)
        http.Error(w, "Failed to update user field", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "user": updatedUser,
    })
}