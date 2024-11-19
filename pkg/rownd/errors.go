package rownd

import "fmt"

// ErrType represents the type of error that occurred
type ErrType string

const (
    ErrAuthentication ErrType = "authentication_error"
    ErrValidation    ErrType = "validation_error"
    ErrAPI           ErrType = "api_error"
    ErrNetwork       ErrType = "network_error"
    ErrNotFound      ErrType = "not_found_error"
)

// RowndError represents a custom error type for Rownd SDK
type RowndError struct {
    Type    ErrType
    Message string
    Err     error
}

func (e *RowndError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Err.Error())
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewError creates a new RowndError
func NewError(errType ErrType, message string, err error) *RowndError {
    return &RowndError{
        Type:    errType,
        Message: message,
        Err:     err,
    }
}