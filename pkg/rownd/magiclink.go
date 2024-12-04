package rownd

import (
	"context"
	"fmt"
	"net/http"
)

// VerificationType determines the method by which this magic link will be verified by the user.
// When the magic link is visited, the user's email or phone number will be marked as verified depending on
// verification mode. In practice, you should set this value to 'phone' if you are
// sending this link to the user via SMS. Likewise, set to 'email' if you are sending
// the magic link to the user via email.
type VerificationType string

func (vt VerificationType) validate() bool {
	switch vt {
	case VerificationTypeEmail, VerificationTypePhone:
		return true
	default:
		return false
	}
}

const (
	VerificationTypeEmail VerificationType = "email"
	VerificationTypePhone VerificationType = "phone"
)

// Purpose determines whether this link will identify a user and allow them to sign in
// automatically or whether it will just act as a simple redirect. An auth
// link should only be sent to a single user, whereas a shorten link can
// be sent to many users if desired, since it does not carry an authentication
// payload.
type Purpose string

const (
	PurposeAuth    Purpose = "auth"
	PurposeShorten Purpose = "shorten"
)

func (p Purpose) validate() bool {
	switch p {
	case PurposeAuth, PurposeShorten:
		return true
	default:
		return false
	}
}

type magicLinkClient struct {
	*Client
}

// CreateMagicLinkRequest ...
type CreateMagicLinkRequest struct {
	// Purpose determines whether this link will identify a user and allow them to sign in
	// automatically or whether it will just act as a simple redirect. An auth
	// link should only be sent to a single user, whereas a shorten link can
	// be sent to many users if desired, since it does not carry an authentication
	// payload.
	Purpose Purpose `json:"purpose"`

	// VerificationType is the means by which this magic link will be verified by the user.
	// When the magic link is visited, the user's email or phone number
	// will be marked as verified depending on
	// verification mode. In practice, you should set this value to 'phone' if you are
	// sending this link to the user via SMS. Likewise, set to 'email' if you are sending
	// the magic link to the user via email.
	VerificationType VerificationType `json:"verification_type"`

	// Data to add to the user's profile. These properties must exist in the Profile Data
	// portion of the Rownd application. If properties such as email, phone, etc identify
	// an existing user, the user will be auto-matched regardless of the user_id value.
	Data map[string]any `json:"data"`

	// RedirectURL is the absolute URL or relative path to send your user after sign-in.
	// If the URL is relative, it will be appended to your application's
	// default redirect URL as defined in your Rownd application settings.
	// If no default redirect is set, the magic link creation will fail.
	RedirectURL string `json:"redirect_url"`

	// Specify a user ID. If the user already exists, include their user ID. Otherwise,
	// use one of '__default__', '__uuid__', or '__objectid__'. These special values
	// tell Rownd to generate an ID in the provided format, or use the application's
	// default user ID format.
	UserID string `json:"user_id,omitempty"`

	// A human-readable string representing the duration for which a magic link is valid.
	// Examples of valid values include 1h, 2d, 3w, 1m, etc. May not exceed 30d.
	// Defaults to 30d.
	Expiration string `json:"expiration,omitempty"` // e.g. "30d"

	// The ID of a group which the user will auto-join upon completing sign-in. The group
	// must have an 'open' admission policy.
	GroupToJoin string `json:"group_to_join,omitempty"`
}

func (r CreateMagicLinkRequest) validate() error {
	var errs []error

	if r.RedirectURL == "" {
		errs = append(errs, NewError(ErrValidation, "redirect_url is required", nil))
	}
	if !r.Purpose.validate() {
		errs = append(errs, NewError(ErrValidation, "purpose is required", nil))
	}
	if !r.VerificationType.validate() {
		errs = append(errs, NewError(ErrValidation, "verification_type is required", nil))
	}
	if r.Data == nil || (r.VerificationType == "email" && r.Data["email"] == "") {
		errs = append(errs, NewError(ErrValidation, "data.email is required for email verification", nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

// MagicLink ...
type MagicLink struct {
	// The magic link URL.
	Link string `json:"link"`
	// The user ID specified during creation or the resolved user ID if one of the directives
	// was specified (e.g. '__default__')
	// TODO: add constants for special App user id values.
	AppUserID string `json:"app_user_id"`
}

// Create creates a new magic link.
func (c *magicLinkClient) Create(ctx context.Context, request CreateMagicLinkRequest) (*MagicLink, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	endpoint, err := c.rowndURL("hub", "auth", "magic")
	if err != nil {
		return nil, err
	}

	var response *MagicLink
	if err := c.request(ctx, http.MethodPost, endpoint.String(), request, &response, c.httpClientOpts...); err != nil {
		return nil, err
	}

	return response, nil
}

// MagicLinkOptions represents the options for creating a magic link
type MagicLinkOptions struct {
	Purpose          string                 `json:"purpose"`
	VerificationType string                 `json:"verification_type"`
	Data             map[string]interface{} `json:"data"`
	RedirectURL      string                 `json:"redirect_url"`
	Expiration       string                 `json:"expiration"`
}

// CreateMagicLink creates a new magic link
func (c *magicLinkClient) CreateMagicLink(ctx context.Context, opts *MagicLinkOptions) (*MagicLink, error) {
	endpoint, err := c.rowndURL("hub", "smart-links")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	var response MagicLink
	if err := c.request(ctx, http.MethodPost, endpoint.String(), opts, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
