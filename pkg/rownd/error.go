package rownd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrKind represents the type of error that occurred
type ErrKind string

const (
	ErrAuthentication ErrKind = "authentication_error"
	ErrValidation     ErrKind = "validation_error"
	ErrAPI            ErrKind = "api_error"
	ErrNetwork        ErrKind = "network_error"
	ErrNotFound       ErrKind = "not_found_error"
)

// Error represents a custom error type for Rownd SDK
type Error struct {
	Kind    ErrKind
	Message string
	Err     error
}

// NewError creates a new RowndError
func NewError(kind ErrKind, message string, err error) *Error {
	return &Error{
		Kind:    kind,
		Message: message,
		Err:     err,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Kind, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// Unwrap implements errors.Unwrap.
func (e *Error) Unwrap() error {
	if e.Err == nil {
		return nil
	}
	return e.Err
}

type MultiError struct {
	errors []error
}

func (e *MultiError) Error() string {
	switch len(e.errors) {
	case 0:
		return ""
	case 1:
		return e.errors[0].Error()
	}

	var b strings.Builder
	for _, err := range e.errors {
		b.WriteString(err.Error())
		b.WriteString("\n")
	}

	return b.String()
}

// ErrorResponse ...
type ErrorResponse struct {
	StatusCode   int      `json:"statusCode"`
	Status       string   `json:"name"`
	ErrorMessage string   `json:"error"`
	Messages     []string `json:"messages"`
}

// Error implements the error interface.
func (er *ErrorResponse) Error() string {
	return er.ErrorMessage
}

// handleErrorResponse ...
func handleErrorResponse(response *http.Response) error {
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	defer response.Body.Close()

	var errorResponse *ErrorResponse
	if err := json.Unmarshal(responseBody, &errorResponse); err != nil {
		return NewError(ErrAPI, fmt.Sprintf("request failed with status %d", response.StatusCode), err)
	}

	return errorResponse
}
