package rownd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// User represents a Rownd user
type User struct {
	ID            string                    `json:"rownd_user"`
	State         string                    `json:"state"`
	AuthLevel     AuthLevel                 `json:"auth_level"`
	Data          map[string]any            `json:"data"`
	VerifiedData  map[string]any            `json:"verified_data"`
	Groups        []UserGroupMembership     `json:"groups"`
	Meta          UserMeta                  `json:"meta"`
	ConnectionMap map[string]UserConnection `json:"connection_map"`
}

// UserConnection ...
type UserConnection struct {
	ConnectionRecordID string                         `json:"connection_record_id"`
	Fields             map[string]UserConnectionField `json:"fields"`
}

// UserConnectionField ...
type UserConnectionField struct {
	InSync bool `json:"in_sync"`
}

// UserMeta ...
type UserMeta struct {
	Created                       time.Time `json:"created"`
	Modified                      time.Time `json:"modified"`
	FirstSignIn                   time.Time `json:"first_sign_in"`
	FirstSignInMethod             string    `json:"first_sign_in_method"`
	LastSignIn                    time.Time `json:"last_sign_in"`
	LastSignInMethod              string    `json:"last_sign_in_method"`
	LastActive                    time.Time `json:"last_active"`
	LastPasskeyRegistrationPrompt time.Time `json:"last_passkey_registration_prompt"`
}

// UserGroupMembership ...
type UserGroupMembership struct {
	Group  Group       `json:"group"`
	Member GroupMember `json:"member"`
}

type userClient struct {
	*Client
}

// GetUserRequest ...
type GetUserRequest struct {
	// UserID is the user id.
	UserID string `json:"-"`
	// Fields is a comma-separated list of fields to include in the profile data.
	Fields []string `json:"-"`
}

func (r *GetUserRequest) params() url.Values {
	q := url.Values{}

	if len(r.Fields) > 0 {
		q.Add("fields", strings.Join(r.Fields, ","))
	}

	return q
}

func (r *GetUserRequest) validate() error {
	var errs []error

	if r.UserID == "" {
		errs = append(errs, NewError(ErrValidation, "user id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Get retrieves a User Profile.
func (c *userClient) Get(ctx context.Context, request GetUserRequest) (*User, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "users", request.UserID, "data")
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response *User
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	// set the ID from the input userID since it's not in the response
	response.ID = request.UserID

	return response, nil
}

// ListUsersRequest ...
type ListUsersRequest struct {
	// Fields is a comma-separated list of fields to include in the profile data
	Fields []string `json:"fields"`

	// LookupFilter filters the resources that match this filter
	// Example: "user@example.com"
	LookupFilter []string `json:"lookup_filter"`

	// A comma-separated list of resource IDs to filter by.
	// TODO does this exclude/include the id?
	IDFilter []string `json:"id_filter"`

	// Number of resources to return per query. Max is 1000.
	PageSize *int `json:"page_size"`

	// ID of the last resource in the previous page. If provided, the next page of results is
	// returned beginning with this resource ID.
	After *string `json:"after"`

	// Sort determines which direction to sort the results
	Sort *Sort `json:"sort"`

	// Include multiple users if they are found using the lookup and id filter
	IncludeDuplicates *bool `json:"include_duplicates"`
}

func (r *ListUsersRequest) params() url.Values {
	q := url.Values{}

	if len(r.Fields) > 0 {
		q.Add("fields", strings.Join(r.Fields, ","))
	}
	if len(r.LookupFilter) > 0 {
		q.Add("lookup_filter", strings.Join(r.LookupFilter, ","))
	}
	if len(r.IDFilter) > 0 {
		q.Add("id_filter", strings.Join(r.IDFilter, ","))
	}
	if r.PageSize != nil {
		q.Add("page_size", strconv.Itoa(*r.PageSize))
	}
	if r.After != nil {
		q.Add("after", *r.After)
	}
	if r.Sort != nil {
		q.Add("sort", string(*r.Sort))
	}

	return q
}

func (r *ListUsersRequest) validate() error {
	var errs []error

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// ListUsersResponse ...
type ListUsersResponse struct {
	// The total number of results.
	TotalResults int `json:"total_results"`
	// The list of user profiles.
	Results []User `json:"results"`
}

// List lists all users for an application.
func (c *userClient) List(ctx context.Context, request ListUsersRequest) (*ListUsersResponse, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "users", "data")
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response *ListUsersResponse
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	// Ensure IDs are properly set for each user in the results
	for i := range response.Results {
		if response.Results[i].ID == "" {
			response.Results[i].ID = response.Results[i].GetID()
		}
	}

	return response, nil
}

// CreateOrUpdateUserRequest represents the request body for updating a user
type CreateOrUpdateUserRequest struct {
	// Rownd User id.
	UserID string `json:"-"`

	// WriteDataToIntegrations is a query parameter that dictates if Rownd should write the
	// profile data changes to integrations attached to your application.
	// default: true
	WriteDataToIntegrations *bool `json:"-"`

	// Data needs to be wrapped in a data object
	Data map[string]interface{} `json:"data"`

	// ConnectionMap is optional
	ConnectionMap []interface{} `json:"connection_map,omitempty"`
}

func (r *CreateOrUpdateUserRequest) params() url.Values {
	q := url.Values{}

	if r.WriteDataToIntegrations != nil {
		q.Add("write_data_to_integrations", strconv.FormatBool(ToValue(r.WriteDataToIntegrations)))
	}

	return q
}

func (r *CreateOrUpdateUserRequest) validate() error {
	var errs []error

	if r.Data == nil {
		errs = append(errs, NewError(ErrValidation, "data is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// CreateOrUpdate updates a user. If UserID is empty, a new user will be created with the supplied information.
func (c *userClient) CreateOrUpdate(ctx context.Context, request CreateOrUpdateUserRequest) (*User, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "users", request.UserID, "data")
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response *User
	if err := c.request(ctx, http.MethodPut, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	// For new user creation, get the ID from data.user_id
	if request.UserID == "__UUID__" && response.ID == "" {
		if userID, ok := response.Data["user_id"].(string); ok {
			response.ID = userID
		}
	} else {
		response.ID = request.UserID
	}

	return response, nil
}

// PatchUserRequest ...
type PatchUserRequest struct {
	UserID string `json:"-"`

	// WriteDataToIntegrations is a query parameter that dictates if Rownd should write the
	// profile data changes to integrations attached to your application.
	// default: true
	WriteDataToIntegrations *bool `json:"-"`

	Data map[string]any `json:"data"`
}

func (r *PatchUserRequest) params() url.Values {
	q := url.Values{}

	if r.WriteDataToIntegrations != nil {
		q.Add("write_data_to_integrations", strconv.FormatBool(ToValue(r.WriteDataToIntegrations)))
	}

	return q
}

func (r *PatchUserRequest) validate() error {
	var errs []error

	if r.UserID == "" {
		errs = append(errs, NewError(ErrValidation, "user id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Patch patches an existing user.
func (c *userClient) Patch(ctx context.Context, request PatchUserRequest) (*User, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "users", request.UserID, "data")
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response *User
	if err := c.request(ctx, http.MethodPatch, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	response.ID = request.UserID

	return response, nil
}

// DeleteUserRequest ...
type DeleteUserRequest struct {
	UserID string `json:"-"`
}

func (r *DeleteUserRequest) validate() error {
	var errs []error

	if r.UserID == "" {
		errs = append(errs, NewError(ErrValidation, "user id is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Delete deletes an existing user and all associated data.
func (c *userClient) Delete(ctx context.Context, request DeleteUserRequest) error {
	if err := request.validate(); err != nil {
		return err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "users", request.UserID, "data")
	if err != nil {
		return err
	}

	if err := c.request(ctx, http.MethodDelete, endpoint.String(), nil, nil, c.httpClientOpts...); err != nil {
		return err
	}

	return nil
}

// GetID returns the user ID from either field
func (u *User) GetID() string {
	if u.ID != "" {
		return u.ID
	}
	// Check data field as fallback
	if userID, ok := u.Data["user_id"].(string); ok {
		return userID
	}
	return ""
}
