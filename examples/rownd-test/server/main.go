package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rownd/client-go/pkg/rownd"
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

type contextKey string

const validationContextKey contextKey = "validation"

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
	// Initialize client with options instead of config struct
	var err error
	client, err = rownd.NewClient(
		rownd.WithAppKey(os.Getenv("ROWND_APP_KEY")),
		rownd.WithAppSecret(os.Getenv("ROWND_APP_SECRET")),
		rownd.WithBaseURL("https://api.rownd.io"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files
	replaceFunc := func(content string) string {
		return strings.ReplaceAll(content, "ROWND_APP_KEY", os.Getenv("ROWND_APP_KEY"))
	}

	// File server for serving static files
	fs := http.FileServer(http.Dir("client/static"))

	// Custom handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get the file path
		filePath := "client/static" + r.URL.Path
		if r.URL.Path == "" || r.URL.Path == "/" {
			filePath += "index.html"
		}

		// Check if the requested file exists
		if _, err := os.Stat(filePath); err == nil {
			// Read the file content
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}

			// Apply the string replacement logic
			modifiedContent := replaceFunc(string(data))

			// Write the modified content to the response
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html") // Adjust the content type as needed
			w.Write([]byte(modifiedContent))
		} else {
			// Serve a 404 page or delegate to the file server
			fs.ServeHTTP(w, r)
		}
	})

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
			// Try both header names
			token := r.Header.Get("Authorization")
			if token == "" {
				token = r.Header.Get("Authentication")
			}
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

			ctx := context.WithValue(r.Context(), validationContextKey, validation)
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
	validation := ctx.Value(validationContextKey).(*rownd.Token)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(validation); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	validation := ctx.Value(validationContextKey).(*rownd.Token)

	// Use new Users.Get method
	user, err := client.Users.Get(ctx, rownd.GetUserRequest{
		UserID: validation.UserID,
	})
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

	// Use new Groups.List method
	groups, err := client.Groups.List(ctx, rownd.ListGroupsRequest{})
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
		TotalResults int                `json:"total_results"`
		Results      []GroupWithMembers `json:"results"`
	}

	// Get members for each group using new GroupMembers.List method
	groupsWithMembers := []GroupWithMembers{}
	for _, group := range groups.Results {
		members, err := client.GroupMembers.List(ctx, rownd.ListGroupMembersRequest{
			GroupID: group.ID,
		})
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

	var req rownd.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use new Groups.Create method
	group, err := client.Groups.Create(r.Context(), req)
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

	validation := r.Context().Value(validationContextKey).(*rownd.Token)

	groupID := r.URL.Query().Get("group_id")
	if groupID == "" {
		http.Error(w, "group_id is required", http.StatusBadRequest)
		return
	}

	var req rownd.CreateGroupInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body provided, use defaults
		req = rownd.CreateGroupInviteRequest{
			UserID:      validation.UserID,
			Roles:       []string{"member"},
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
	invite, err := client.GroupInvites.Create(ctx, rownd.CreateGroupInviteRequest{
		GroupID:     groupID,
		UserID:      req.UserID,
		Roles:       req.Roles,
		RedirectURL: req.RedirectURL,
	})
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
	validation := ctx.Value(validationContextKey).(*rownd.Token)

	field := strings.TrimPrefix(r.URL.Path, "/user/field/")
	if field == "" {
		http.Error(w, "Field name is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Use UserFields.Update instead of UpdateField
	err := client.UserFields.Update(ctx, rownd.UpdateUserFieldRequest{
		UserID: validation.UserID,
		Field:  field,
		Value:  req.Value,
	})
	if err != nil {
		log.Printf("Error updating user field: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to update user field", err.Error())
		return
	}

	// Fetch updated user data
	user, err := client.Users.Get(ctx, rownd.GetUserRequest{
		UserID: validation.UserID,
	})
	if err != nil {
		log.Printf("Error fetching updated user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch updated user", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    user,
	})
}
