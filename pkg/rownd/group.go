package rownd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// AdmissionPolicy determines whether the group is open for anyone to join or by invite only.
type AdmissionPolicy string

const (
	AdmissionPolicyInviteOnly AdmissionPolicy = "invite_only"
	AdmissionPolicyOpen       AdmissionPolicy = "open"
)

func (ap AdmissionPolicy) validate() bool {
	switch ap {
	case AdmissionPolicyInviteOnly, AdmissionPolicyOpen:
		return true
	default:
		return false
	}
}

// Group ...
type Group struct {
	// ID is the group ID.
	ID string `json:"id"`
	// Name is the group name.
	Name string `json:"name"`
	// MemberCount is the number of members in the group.
	// This value is no longer provided by the API. The default value of 0 will always be returned
	MemberCount int `json:"member_count"`
	// AppID is the Rownd application ID.
	AppID string `json:"app_id"`
	// AdmissionPolicy indicates if the group is open or requires an invitation to join.
	AdmissionPolicy AdmissionPolicy `json:"admission_policy"`
	// Meta is an object containing additional metadata for the group
	Meta map[string]any `json:"meta,omitempty"`
	// CreatedBy is the ID of the user that created the resource.
	CreatedBy string `json:"created_by"`
	// CreatedAt is the ISO 8601 date-time that the resource was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedBy is the ID of the user that most recently updated the resource.
	UpdatedBy string `json:"updated_by"`
	// UpdatedAt is the ISO 8601 date-time that the resource was updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// groupClient ...
type groupClient struct {
	*Client
}

// GetGroupRequest ...
type GetGroupRequest struct {
	GroupID string `json:"-"`
}

// validate ...
func (r GetGroupRequest) validate() error {
	var errs []error

	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Get retrieves a specific group.
func (c *groupClient) Get(ctx context.Context, request GetGroupRequest) (*Group, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL(c.baseURL, "applications", c.appID, "groups", request.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to compose endpoint: %w", err)
	}

	var response *Group
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// ListGroupsRequest ...
type ListGroupsRequest struct {
	// PageSize is the number of resources to return per query. Max is 100.
	PageSize *int

	// After is the ID of the last resource in the previous page. If provided, the next page of results is
	// returned beginning with this resource ID.
	After *string

	// LookupFilter return resources that match this filter
	LookupFilter []string
}

func (r ListGroupsRequest) params() url.Values {
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

func (r ListGroupsRequest) validate() error {
	var errs []error

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// ListGroupsResponse ...
type ListGroupsResponse struct {
	TotalResults int     `json:"total_results"`
	Results      []Group `json:"results"`
}

// List retrieves all groups for an application.
func (c *groupClient) List(ctx context.Context, request ListGroupsRequest) (*ListGroupsResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "groups")
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response *ListGroupsResponse
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// CreateGroupRequest ...
type CreateGroupRequest struct {
	// The group name.
	Name string `json:"name"`

	// AdmissionPolicy sets whether the group is open for anyone to join or by invite only.
	AdmissionPolicy AdmissionPolicy `json:"admission_policy"`

	// Meta is an object containing additional metadata for the group.
	Meta map[string]any `json:"meta,omitempty"`
}

func (r CreateGroupRequest) validate() error {
	var errs []error

	if !r.AdmissionPolicy.validate() {
		errs = append(errs, NewError(ErrValidation, "invalid admission policy", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Create creates a new group.
func (c *groupClient) Create(ctx context.Context, request CreateGroupRequest) (*Group, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "groups")
	if err != nil {
		return nil, err
	}

	var response *Group
	if err := c.request(ctx, http.MethodPost, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// DeleteGroupRequest ...
type DeleteGroupRequest struct {
	GroupID string
}

func (r *DeleteGroupRequest) validate() error {
	var errs []error

	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Delete removes a group
func (c *groupClient) Delete(ctx context.Context, req DeleteGroupRequest) error {
	if err := req.validate(); err != nil {
		return err
	}
	endpoint, err := c.rowndURL("applications", c.appID, "groups", req.GroupID)
	if err != nil {
		return err
	}

	return c.request(ctx, http.MethodDelete, endpoint.String(), nil, nil, c.httpClientOpts...)
}
