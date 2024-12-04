package rownd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type userFieldClient struct {
	*Client
}

// GetUserFieldRequest ...
type GetUserFieldRequest struct {
	UserID string
	Field  string

	Preview     *bool
	FailOnError *bool
}

func (r GetUserFieldRequest) params() url.Values {
	q := url.Values{}

	if r.Preview != nil {
		q.Add("preview", strconv.FormatBool(ToValue(r.Preview)))
	}
	if r.FailOnError != nil {
		q.Add("fail_on_error", strconv.FormatBool(ToValue(r.FailOnError)))
	}

	return q
}

func (r GetUserFieldRequest) validate() error {
	var errs []error

	if r.UserID == "" {
		errs = append(errs, NewError(ErrValidation, "user id is required", nil))
	}
	if r.Field == "" {
		errs = append(errs, NewError(ErrValidation, "field is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Get retrieves an existing user field.
func (c *userFieldClient) Get(ctx context.Context, request GetUserFieldRequest) (any, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "users", request.UserID, "data", "fields", request.Field)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = request.params().Encode()

	var response map[string]any
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	value, ok := response["value"]
	if !ok {
		return nil, fmt.Errorf("value for field %s not found", request.Field)
	}

	return value, nil
}

// UpdateUserFieldRequest ...
type UpdateUserFieldRequest struct {
	UserID string `json:"-"`
	Field  string `json:"-"`
	Value  any    `json:"value"`
}

func (r UpdateUserFieldRequest) validate() error {
	var errs []error

	if r.UserID == "" {
		errs = append(errs, NewError(ErrValidation, "user id is required", nil))
	}
	if r.Field == "" {
		errs = append(errs, NewError(ErrValidation, "field is required", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// Update updates an existing user field.
func (c *userFieldClient) Update(ctx context.Context, request UpdateUserFieldRequest) error {
	if err := request.validate(); err != nil {
		return err
	}

	endpoint, err := c.rowndURL("applications", c.appID, "users", request.UserID, "data", "fields", request.Field)
	if err != nil {
		return err
	}

	fmt.Printf("Making field update request to: %s\n", endpoint.String())

	if err := c.request(ctx, http.MethodPut, endpoint.String(), request, nil, c.httpClientOpts...); err != nil {
		fmt.Printf("Field update error: %v\n", err)
		return err
	}

	return nil
}
