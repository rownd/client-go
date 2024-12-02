package rownd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// UserListOptions represents the options for listing/filtering users
type UserListOptions struct {
	LookupValue       string   `json:"lookup_filter,omitempty"`     // The value to search for (email/phone)
	Fields            []string `json:"fields,omitempty"`            // Fields to include in response
	IDFilter          []string `json:"id_filter,omitempty"`         // List of user IDs to filter by
	PageSize          int      `json:"page_size,omitempty"`         // Number of results per page (max 1000)
	After             string   `json:"after,omitempty"`             // ID of last resource from previous page
	Sort              string   `json:"sort,omitempty"`              // Sort direction: "asc" or "desc"
	IncludeDuplicates bool     `json:"include_duplicates,omitempty"` // Include multiple matches
}

func (c *Client) GetUser(ctx context.Context, userID string, tokenInfo *TokenValidationResponse) (*User, error) {
	// Get app ID from token claims
	var appID string
	if tokenInfo != nil && tokenInfo.DecodedToken != nil {
		if aud, exists := tokenInfo.DecodedToken["aud"]; exists {
			switch v := aud.(type) {
			case []interface{}:
				if len(v) > 0 {
					if audStr, ok := v[0].(string); ok && strings.HasPrefix(audStr, "app:") {
						appID = audStr[4:]
					}
				}
			case []string:
				if len(v) > 0 && strings.HasPrefix(v[0], "app:") {
					appID = v[0][4:]
				}
			case string:
				if strings.HasPrefix(v, "app:") {
					appID = v[4:]
				}
			}
		}
	}

	// If no app ID in token, use the one from client config
	if appID == "" {
		appID = c.AppID
	}

	if appID == "" {
		return nil, fmt.Errorf("app ID not found in token or client config")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/applications/%s/users/%s/data", c.BaseURL, appID, userID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and log the response body for debugging
	respBody, err := io.ReadAll(resp.Body)
	fmt.Printf("\nGet User Response Body: %s\n", string(respBody))
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	if resp.StatusCode != http.StatusOK {
		var apiErr APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s", apiErr.Error)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	// Set the ID from the input userID since it's not in the response
	user.ID = userID

	return &user, nil
}

func (c *Client) UpdateUser(ctx context.Context, appID string, userID string, userData map[string]interface{}) (*User, error) {
	if appID == "" {
		return nil, fmt.Errorf("app ID is required")
	}

	// For new users without ID, use __UUID__ to have Rownd generate one
	if userID == "" {
		userID = "__UUID__"
	}

	endpoint := fmt.Sprintf("%s/applications/%s/users/%s/data", c.BaseURL, appID, userID)
	fmt.Printf("\nAPI Request URL: %s\n", endpoint)

	// Use the userData directly if it has a "data" wrapper, otherwise wrap it
	var payload map[string]interface{}
	if _, hasData := userData["data"]; hasData {
		payload = userData
	} else {
		payload = map[string]interface{}{
			"data": userData,
		}
	}

	// Log request details
	payloadBytes, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Printf("Request Headers:\n")
	fmt.Printf("  x-rownd-app-key: %s\n", c.AppKey)
	fmt.Printf("  x-rownd-app-secret: %s...\n", c.AppSecret[:10])
	fmt.Printf("Request Payload:\n%s\n", string(payloadBytes))

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, jsonReader(payload))
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	return c.doUserRequest(req)
}

func (c *Client) PatchUser(ctx context.Context, appID string, userID string, data map[string]interface{}) (*User, error) {
	if appID == "" {
		return nil, fmt.Errorf("app ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	payload := map[string]interface{}{
		"data": data,
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH",
		fmt.Sprintf("%s/applications/%s/users/%s/data", c.BaseURL, appID, userID),
		jsonReader(payload))
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	return c.doUserRequest(req)
}

func (c *Client) GetUserField(ctx context.Context, appID string, userID string, field string) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/applications/%s/users/%s/data/fields/%s", c.BaseURL, appID, userID, field),
		nil)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, NewError(ErrAPI, "failed to decode response", err)
	}

	return result["value"], nil
}

func (c *Client) UpdateUserField(ctx context.Context, appID string, userID string, field string, value interface{}) error {
	payload := map[string]interface{}{
		"value": value,
	}

	req, err := http.NewRequestWithContext(ctx, "PUT",
		fmt.Sprintf("%s/applications/%s/users/%s/data/fields/%s", c.BaseURL, appID, userID, field),
		jsonReader(payload))
	if err != nil {
		return NewError(ErrAPI, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)

	_, err = c.doRequest(req)
	return err
}

func jsonReader(v interface{}) io.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}

func (c *Client) doUserRequest(req *http.Request) (*User, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return nil, NewError(ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	// Read and log the response body
	respBody, err := io.ReadAll(resp.Body)
	fmt.Printf("\nResponse Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n", string(respBody))

	// Create new reader from the response body for further processing
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, NewError(ErrAPI, "failed to decode response", err)
	}

	// Set the ID from the user_id in the data field
	if userID, ok := user.Data["user_id"].(string); ok {
		user.ID = userID
	}

	return &user, nil
}

// DeleteUser deletes a user and all associated data
func (c *Client) DeleteUser(ctx context.Context, appID string, userID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE",
		fmt.Sprintf("%s/applications/%s/users/%s/data", c.BaseURL, appID, userID),
		nil)
	if err != nil {
		return NewError(ErrAPI, "failed to create request", err)
	}

	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return NewError(ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	// 204 No Content is a success response for DELETE
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return handleErrorResponse(resp)
	}

	return nil
}

// LookupUsers searches for users based on the provided options
func (c *Client) LookupUsers(ctx context.Context, appID string, opts *UserListOptions) (*UserListResponse, error) {
	if appID == "" {
		return nil, NewError(ErrValidation, "app ID is required", nil)
	}

	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/applications/%s/users/data", c.BaseURL, appID)
	if opts.LookupValue != "" {
		baseURL = fmt.Sprintf("%s?lookup_filter=%s", baseURL, url.QueryEscape(opts.LookupValue))
	}

	fmt.Printf("\nLookup Request URL: %s\n", baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, NewError(ErrAPI, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-rownd-app-key", c.AppKey)
	req.Header.Set("x-rownd-app-secret", c.AppSecret)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, NewError(ErrNetwork, "request failed", err)
	}
	defer resp.Body.Close()

	// Read and log response
	respBody, err := io.ReadAll(resp.Body)
	fmt.Printf("Lookup Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Lookup Response Body: %s\n", string(respBody))
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	var listResp UserListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, NewError(ErrAPI, "failed to decode response", err)
	}

	// Set the ID for each user from their data
	for i := range listResp.Results {
		if userID, ok := listResp.Results[i].Data["user_id"].(string); ok {
			listResp.Results[i].ID = userID
		}
	}

	return &listResp, nil
}
