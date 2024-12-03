package rownd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// GroupInvite ...
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

type groupInviteClient struct {
	*Client
}

// GetGroupInviteRequest ...
type GetGroupInviteRequest struct {
	AppID    string
	GroupID  string
	InviteID string
}

func (r GetGroupInviteRequest) validate() error {
	var errs []error

	if r.AppID == "" {
		errs = append(errs, NewError(ErrValidation, "app id is required", nil))
	}
	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}
	if r.InviteID == "" {
		errs = append(errs, NewError(ErrValidation, "invite id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Get retrieves a specific group invite
func (c *groupInviteClient) Get(ctx context.Context, request GetGroupInviteRequest) (*GroupInvite, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", request.AppID, "groups", request.GroupID, "invites", request.InviteID)
	if err != nil {
		return nil, err
	}

	var response *GroupInvite
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// ListGroupInvitesRequest ...
type ListGroupInvitesRequest struct {
	AppID   string
	GroupID string

	// EnsuredUserID is the User ID for which the invite was created. This is not the member ID.
	EnsuredUserID *string
}

func (r ListGroupInvitesRequest) params() url.Values {
	q := url.Values{}

	if r.EnsuredUserID != nil {
		q.Add("ensured_user_id", ToValue(r.EnsuredUserID))
	}

	return q
}

func (r ListGroupInvitesRequest) validate() error {
	var errs []error

	if r.AppID == "" {
		errs = append(errs, NewError(ErrValidation, "app id is required", nil))
	}
	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// ListGroupInvitesResponse ...
type ListGroupInvitesResponse struct {
	TotalResults int           `json:"total_results"`
	Results      []GroupInvite `json:"results"`
}

// List lists all invites for a group
func (c *groupInviteClient) List(ctx context.Context, request ListGroupInvitesRequest) (*ListGroupInvitesResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", request.AppID, "groups", request.GroupID, "invites")
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response *ListGroupInvitesResponse
	if err := c.request(ctx, http.MethodDelete, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

type CreateGroupInviteRequest struct {
	// AppID is the Rownd application ID
	AppID string `json:"-"`

	// GroupID is Group ID.
	GroupID string `json:"-"`

	// UserID is the ID of a Rownd user in the specified application.
	UserID string `json:"user_id,omitempty"`

	// Roles are the roles into which a group member will be added upon invite acceptance (The first
	// member invited to a group will always be created with the 'owner' role along with any
	// additional roles specified)
	Roles []string `json:"roles"`

	// RedirectURL is the relative or absolute path location to which a user
	// will be directed after accepting the invite.
	RedirectURL string `json:"redirect_url,omitempty"`

	// Email is the email of a Rownd user in the specified application.
	// This property is mutually exclusive with user_id and email.
	Email string `json:"email,omitempty"`

	// Phone is the phone number of a Rownd user in the specified application.
	// This property is mutually exclusive with user_id and email
	Phone int64 `json:"phone,omitempty"`

	// AppVariantID is the ID of an application variant for which this invite should be created.
	// When a user accepts this invite, they will be added to the group as a
	// member and signed in to this application variant.
	AppVariantID string `json:"app_variant_id,omitempty"`
}

func (r CreateGroupInviteRequest) validate() error {
	var errs []error

	if r.AppID == "" {
		errs = append(errs, NewError(ErrValidation, "app id is required", nil))
	}
	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}
	if len(r.Roles) == 0 {
		errs = append(errs, NewError(ErrValidation, "roles is required", nil))
	}
	// TODO add validation rules for mutually exclusives fields.

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// GroupInviteResponse ...
type GroupInviteResponse struct {
	// Link is the invitation link. Your user will use this link to accept the invite.
	// For now, you will need to send this link to your users. In the future, Rownd will
	// automatically send out invites via email or SMS for you.
	Link       string      `json:"link"`
	Invitation GroupInvite `json:"invitation"`
}

// Create creates a new group invite.
func (c *groupInviteClient) Create(ctx context.Context, request CreateGroupInviteRequest) (*GroupInviteResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", request.AppID, "groups", request.GroupID, "invites")
	if err != nil {
		return nil, err
	}

	var response *GroupInviteResponse
	if err := c.request(ctx, http.MethodPost, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// UpdateGroupInviteRequest ...
type UpdateGroupInviteRequest struct {
	AppID    string `json:"-"`
	GroupID  string `json:"-"`
	InviteID string `json:"-"`

	UserID       string   `json:"user_id,omitempty"`
	Roles        []string `json:"roles"`
	RedirectURL  string   `json:"redirect_url,omitempty"`
	Email        string   `json:"email,omitempty"`
	Phone        int64    `json:"phone,omitempty"`
	AppVariantID string   `json:"app_variant_id,omitempty"`
}

func (r *UpdateGroupInviteRequest) validate() error {
	var errs []error

	if r.AppID == "" {
		errs = append(errs, NewError(ErrValidation, "app id is required", nil))
	}
	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}
	if r.InviteID == "" {
		errs = append(errs, NewError(ErrValidation, "invite id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Update updates an existing group invite.
func (c *groupInviteClient) Update(ctx context.Context, request UpdateGroupInviteRequest) (*GroupInvite, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", request.AppID, "groups", request.GroupID, "invites", request.InviteID)
	if err != nil {
		return nil, err
	}

	var response *GroupInvite
	if err := c.request(ctx, http.MethodPut, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// DeleteGroupInviteRequest ...
type DeleteGroupInviteRequest struct {
	AppID    string `json:"-"`
	GroupID  string `json:"-"`
	InviteID string `json:"-"`
}

func (r DeleteGroupInviteRequest) validate() error {
	var errs []error

	if r.AppID == "" {
		errs = append(errs, NewError(ErrValidation, "app id is required", nil))
	}
	if r.GroupID == "" {
		errs = append(errs, NewError(ErrValidation, "group id is required", nil))
	}
	if r.InviteID == "" {
		errs = append(errs, NewError(ErrValidation, "invite id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Delete deletes a group invite.
func (c *groupInviteClient) Delete(ctx context.Context, request DeleteGroupInviteRequest) error {
	if err := request.validate(); err != nil {
		return err
	}

	endpoint, err := c.rowndURL("applications", request.AppID, "groups", request.GroupID, "invites", request.InviteID)
	if err != nil {
		return err
	}

	if err := c.request(ctx, http.MethodDelete, endpoint.String(), nil, nil, c.httpClientOpts...); err != nil {
		return err
	}

	return nil
}
