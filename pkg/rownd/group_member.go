package rownd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Logger interface for logging
type Logger interface {
	Printf(format string, v ...interface{})
}

type groupMemberClient struct {
	*Client
	logger Logger
}

// GroupMember ...
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

// GetGroupMemberRequest ...
type GetGroupMemberRequest struct {
	GroupID  string `json:"-"`
	MemberID string `json:"-"`
}

func (r GetGroupMemberRequest) validate() error {
	var errs []error

	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}
	if r.MemberID == "" {
		errs = append(errs, NewError(ErrValidation, "member id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Get retrieves a specific group member.
func (c *groupMemberClient) Get(ctx context.Context, request GetGroupMemberRequest) (*GroupMember, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "groups", request.GroupID, "members", request.MemberID)
	if err != nil {
		return nil, err
	}

	var response *GroupMember
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// ListGroupMembersRequest ...
type ListGroupMembersRequest struct {
	// Group ID
	GroupID string
	// PageSize is the number of resources to return per query. Max is 100.
	PageSize *int
	// After is the ID of the last resource in the previous page. If provided, the next page of results is
	// returned beginning with this resource ID.
	After *string
	// LookupFilter return resources that match this filter
	LookupFilter []string
}

func (r ListGroupMembersRequest) params() url.Values {
	q := url.Values{}

	if r.PageSize != nil {
		q.Add("page_size", strconv.Itoa(ToValue(r.PageSize)))
	}
	if r.After != nil {
		q.Add("after", ToValue(r.After))
	}
	if r.LookupFilter != nil {
		q.Add("lookup_filter", strings.Join(r.LookupFilter, ","))
	}

	return q
}

func (r ListGroupMembersRequest) validate() error {
	var errs []error

	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// ListGroupMembersResponse ...
type ListGroupMembersResponse struct {
	TotalResults int           `json:"total_results"`
	Results      []GroupMember `json:"results"`
}

// List lists all members in a group.
func (c *groupMemberClient) List(ctx context.Context, request ListGroupMembersRequest) (*ListGroupMembersResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "groups", request.GroupID, "members")
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response *ListGroupMembersResponse
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// CreateGroupMemberRequest ...
type CreateGroupMemberRequest struct {
	GroupID string `json:"-"`

	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	State  string   `json:"state"`
}

func (r CreateGroupMemberRequest) validate() error {
	var errs []error

	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}

	// TODO validate other fields

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Create adds a new member to a group
func (c *groupMemberClient) Create(ctx context.Context, request CreateGroupMemberRequest) (*GroupMember, error) {
	if err := request.validate(); err != nil {
		c.logger.Printf("Validation error: %v", err)
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "groups", request.GroupID, "members")
	if err != nil {
		c.logger.Printf("URL creation error: %v", err)
		return nil, err
	}

	// Log request details
	c.logger.Printf("Creating group member - Group ID: %s, User ID: %s", request.GroupID, request.UserID)
	c.logger.Printf("POST Request URL: %s", endpoint.String())
	c.logger.Printf("Request body: %+v", request)

	var response *GroupMember
	if err := c.request(ctx, http.MethodPost, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		c.logger.Printf("API error: %v", err)
		return nil, err
	}

	c.logger.Printf("Response: %+v", response)
	return response, nil
}

// UpdateGroupMemberRequest ...
type UpdateGroupMemberRequest struct {
	GroupID  string `json:"-"`
	MemberID string `json:"-"`

	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	State  string   `json:"state"`
}

func (r UpdateGroupMemberRequest) validate() error {
	var errs []error

	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}
	if r.MemberID == "" {
		errs = append(errs, NewError(ErrValidation, "member id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Update updates an existing group member
func (c *groupMemberClient) Update(ctx context.Context, request UpdateGroupMemberRequest) (*GroupMember, error) {
	if err := request.validate(); err != nil {
		c.logger.Printf("Validation error: %v", err)
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "groups", request.GroupID, "members", request.MemberID)
	if err != nil {
		c.logger.Printf("URL creation error: %v", err)
		return nil, err
	}

	c.logger.Printf("Updating member - Member ID: %s, New roles: %v", request.MemberID, request.Roles)
	c.logger.Printf("PUT Request URL: %s", endpoint.String())

	var response *GroupMember
	if err := c.request(ctx, http.MethodPut, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		c.logger.Printf("Update error: %v", err)
		return nil, err
	}

	c.logger.Printf("Update response: %+v", response)
	return response, nil
}

// DeleteGroupMemberRequest ...
type DeleteGroupMemberRequest struct {
	GroupID  string `json:"-"`
	MemberID string `json:"-"`
}

func (r DeleteGroupMemberRequest) validate() error {
	var errs []error

	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}
	if r.MemberID == "" {
		errs = append(errs, NewError(ErrValidation, "member id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Delete removes a member from a group
func (c *groupMemberClient) Delete(ctx context.Context, req DeleteGroupMemberRequest) error {
	if err := req.validate(); err != nil {
		return err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "groups", req.GroupID, "members", req.MemberID)
	if err != nil {
		return err
	}

	c.logger.Printf("Deleting group member - Group ID: %s, Member ID: %s", req.GroupID, req.MemberID)
	c.logger.Printf("DELETE Request URL: %s", endpoint.String())

	// Pass nil for the response parameter since DELETE returns no content
	if err := c.request(ctx, http.MethodDelete, endpoint.String(), nil, nil, c.httpClientOpts...); err != nil {
		c.logger.Printf("Delete error: %v", err)
		return err
	}

	return nil
}
