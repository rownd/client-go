package rownd

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Group struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    MemberCount     int                    `json:"member_count"`
    AppID           string                 `json:"app_id"`
    AdmissionPolicy string                 `json:"admission_policy"`
    Meta            map[string]interface{} `json:"meta,omitempty"`
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
    UpdatedBy       string                 `json:"updated_by"`
    CreatedBy       string                 `json:"created_by"`
}

type GroupListResponse struct {
    TotalResults int     `json:"total_results"`
    Results      []Group `json:"results"`
}

type CreateGroupRequest struct {
    Name            string                 `json:"name"`
    AdmissionPolicy string                 `json:"admission_policy"`
    Meta            map[string]interface{} `json:"meta,omitempty"`
}

type GroupInvite struct {
    ID              string    `json:"id"`
    GroupID         string    `json:"group_id"`
    Roles           []string  `json:"roles"`
    State           string    `json:"state"`
    Email           string    `json:"email,omitempty"`
    Phone           int64     `json:"phone,omitempty"`
    UserID          string    `json:"user_id,omitempty"`
    UserLookupValue string    `json:"user_lookup_value,omitempty"`
    RedirectURL     string    `json:"redirect_url,omitempty"`
    AppVariantID    string    `json:"app_variant_id,omitempty"`
    CreatedAt       time.Time `json:"created_at"`
    CreatedBy       string    `json:"created_by"`
    AcceptedBy      string    `json:"accepted_by,omitempty"`
    EnsuredUserID   string    `json:"ensured_user_id,omitempty"`
}

type CreateGroupInviteRequest struct {
    UserID       string   `json:"user_id,omitempty"`
    Roles        []string `json:"roles"`
    RedirectURL  string   `json:"redirect_url,omitempty"`
    Email        string   `json:"email,omitempty"`
    Phone        int64    `json:"phone,omitempty"`
    AppVariantID string   `json:"app_variant_id,omitempty"`
}

type GroupInviteResponse struct {
    Link       string      `json:"link"`
    Invitation GroupInvite `json:"invitation"`
}

type GroupInviteListResponse struct {
    TotalResults int           `json:"total_results"`
    Results      []GroupInvite `json:"results"`
}

type GroupMember struct {
    ID        string                 `json:"id"`
    UserID    string                 `json:"user_id"`
    Roles     []string               `json:"roles"`
    State     string                 `json:"state"`
    InvitedBy string                 `json:"invited_by,omitempty"`
    AddedBy   string                 `json:"added_by,omitempty"`
    Profile   map[string]interface{} `json:"profile,omitempty"`
    GroupID   string                 `json:"group_id"`
}

type GroupMemberListResponse struct {
    TotalResults int           `json:"total_results"`
    Results      []GroupMember `json:"results"`
}

type CreateGroupMemberRequest struct {
    UserID string   `json:"user_id"`
    Roles  []string `json:"roles"`
    State  string   `json:"state"`
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, appID string, req *CreateGroupRequest) (*Group, error) {
    payload, err := json.Marshal(req)
    if err != nil {
        return nil, NewError(ErrValidation, "failed to marshal request", err)
    }

    apiReq, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/applications/%s/groups", c.BaseURL, appID),
        bytes.NewBuffer(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupRequest(apiReq)
}

// GetGroup retrieves a specific group
func (c *Client) GetGroup(ctx context.Context, appID string, groupID string) (*Group, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("%s/applications/%s/groups/%s", c.BaseURL, appID, groupID),
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupRequest(req)
}

// ListGroups retrieves all groups for an application
func (c *Client) ListGroups(ctx context.Context, appID string) (*GroupListResponse, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("%s/applications/%s/groups", c.BaseURL, appID),
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var listResp GroupListResponse
    if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &listResp, nil
}

func (c *Client) doGroupRequest(req *http.Request) (*Group, error) {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var group Group
    if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &group, nil
}

// CreateGroupInvite creates a new group invite
func (c *Client) CreateGroupInvite(ctx context.Context, appID string, groupID string, req *CreateGroupInviteRequest) (*GroupInviteResponse, error) {
    payload, err := json.Marshal(req)
    if err != nil {
        return nil, NewError(ErrValidation, "failed to marshal request", err)
    }

    apiReq, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/applications/%s/groups/%s/invites", c.BaseURL, appID, groupID),
        bytes.NewBuffer(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupInviteResponse(apiReq)
}

// GetGroupInvite retrieves a specific group invite
func (c *Client) GetGroupInvite(ctx context.Context, appID string, groupID string, inviteID string) (*GroupInvite, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("%s/applications/%s/groups/%s/invites/%s", c.BaseURL, appID, groupID, inviteID),
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupInviteRequest(req)
}

// UpdateGroupInvite updates an existing group invite
func (c *Client) UpdateGroupInvite(ctx context.Context, appID string, groupID string, inviteID string, req *CreateGroupInviteRequest) (*GroupInvite, error) {
    payload, err := json.Marshal(req)
    if err != nil {
        return nil, NewError(ErrValidation, "failed to marshal request", err)
    }

    apiReq, err := http.NewRequestWithContext(ctx, "PUT",
        fmt.Sprintf("%s/applications/%s/groups/%s/invites/%s", c.BaseURL, appID, groupID, inviteID),
        bytes.NewBuffer(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupInviteRequest(apiReq)
}

// DeleteGroupInvite deletes a group invite
func (c *Client) DeleteGroupInvite(ctx context.Context, appID string, groupID string, inviteID string) error {
    req, err := http.NewRequestWithContext(ctx, "DELETE",
        fmt.Sprintf("%s/applications/%s/groups/%s/invites/%s", c.BaseURL, appID, groupID, inviteID),
        nil)
    if err != nil {
        return NewError(ErrAPI, "failed to create request", err)
    }

    _, err = c.doRequest(req)
    return err
}

// ListGroupInvites lists all invites for a group
func (c *Client) ListGroupInvites(ctx context.Context, appID string, groupID string) (*GroupInviteListResponse, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("%s/applications/%s/groups/%s/invites", c.BaseURL, appID, groupID),
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var listResp GroupInviteListResponse
    if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &listResp, nil
}

// Helper functions
func (c *Client) doGroupInviteRequest(req *http.Request) (*GroupInvite, error) {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var invite GroupInvite
    if err := json.NewDecoder(resp.Body).Decode(&invite); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &invite, nil
}

func (c *Client) doGroupInviteResponse(req *http.Request) (*GroupInviteResponse, error) {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var inviteResp GroupInviteResponse
    if err := json.NewDecoder(resp.Body).Decode(&inviteResp); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &inviteResp, nil
}

// CreateGroupMember adds a new member to a group
func (c *Client) CreateGroupMember(ctx context.Context, appID string, groupID string, req *CreateGroupMemberRequest) (*GroupMember, error) {
    payload, err := json.Marshal(req)
    if err != nil {
        return nil, NewError(ErrValidation, "failed to marshal request", err)
    }

    apiReq, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/applications/%s/groups/%s/members", c.BaseURL, appID, groupID),
        bytes.NewBuffer(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupMemberRequest(apiReq)
}

// GetGroupMember retrieves a specific group member
func (c *Client) GetGroupMember(ctx context.Context, appID string, groupID string, memberID string) (*GroupMember, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("%s/applications/%s/groups/%s/members/%s", c.BaseURL, appID, groupID, memberID),
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupMemberRequest(req)
}

// UpdateGroupMember updates an existing group member
func (c *Client) UpdateGroupMember(ctx context.Context, appID string, groupID string, memberID string, req *CreateGroupMemberRequest) (*GroupMember, error) {
    payload, err := json.Marshal(req)
    if err != nil {
        return nil, NewError(ErrValidation, "failed to marshal request", err)
    }

    apiReq, err := http.NewRequestWithContext(ctx, "PUT",
        fmt.Sprintf("%s/applications/%s/groups/%s/members/%s", c.BaseURL, appID, groupID, memberID),
        bytes.NewBuffer(payload))
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    return c.doGroupMemberRequest(apiReq)
}

// DeleteGroupMember removes a member from a group
func (c *Client) DeleteGroupMember(ctx context.Context, appID string, groupID string, memberID string) error {
    req, err := http.NewRequestWithContext(ctx, "DELETE",
        fmt.Sprintf("%s/applications/%s/groups/%s/members/%s", c.BaseURL, appID, groupID, memberID),
        nil)
    if err != nil {
        return NewError(ErrAPI, "failed to create request", err)
    }

    _, err = c.doRequest(req)
    return err
}

// ListGroupMembers lists all members in a group
func (c *Client) ListGroupMembers(ctx context.Context, appID string, groupID string) (*GroupMemberListResponse, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("%s/applications/%s/groups/%s/members", c.BaseURL, appID, groupID),
        nil)
    if err != nil {
        return nil, NewError(ErrAPI, "failed to create request", err)
    }

    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var listResp GroupMemberListResponse
    if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &listResp, nil
}

func (c *Client) doGroupMemberRequest(req *http.Request) (*GroupMember, error) {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, handleErrorResponse(resp)
    }

    var member GroupMember
    if err := json.NewDecoder(resp.Body).Decode(&member); err != nil {
        return nil, NewError(ErrAPI, "failed to decode response", err)
    }

    return &member, nil
}

func handleErrorResponse(resp *http.Response) error {
    var errResp struct {
        Message string `json:"message"`
        Code    string `json:"code"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
        return NewError(ErrAPI, fmt.Sprintf("request failed with status %d", resp.StatusCode), err)
    }
    return NewError(ErrAPI, errResp.Message, nil)
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-rownd-app-key", c.AppKey)
    req.Header.Set("x-rownd-app-secret", c.AppSecret)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, NewError(ErrNetwork, "request failed", err)
    }

    if resp.StatusCode != http.StatusOK {
        defer resp.Body.Close()
        return nil, handleErrorResponse(resp)
    }

    return resp, nil
}