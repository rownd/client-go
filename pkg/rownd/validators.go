package rownd

import (
    "net/mail"
    "regexp"
)

// ValidationError represents a validation error
type ValidationError struct {
    Field   string
    Message string
}

// Validator provides validation methods for various inputs
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
    return &Validator{}
}

// ValidateEmail checks if the provided email is valid
func (v *Validator) ValidateEmail(email string) error {
    _, err := mail.ParseAddress(email)
    if err != nil {
        return NewError(ErrValidation, "invalid email format", err)
    }
    return nil
}

// ValidateToken checks if the provided token format is valid
func (v *Validator) ValidateToken(token string) error {
    if token == "" {
        return NewError(ErrValidation, "token cannot be empty", nil)
    }
    
    // Basic JWT format validation
    jwtRegex := regexp.MustCompile(`^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]*$`)
    if !jwtRegex.MatchString(token) {
        return NewError(ErrValidation, "invalid token format", nil)
    }
    
    return nil
}

// ValidateUserID checks if the provided user ID is valid
func (v *Validator) ValidateUserID(userID string) error {
    if userID == "" {
        return NewError(ErrValidation, "user ID cannot be empty", nil)
    }
    
    // Add any specific user ID format validation rules here
    if len(userID) < 3 {
        return NewError(ErrValidation, "user ID must be at least 3 characters long", nil)
    }
    
    return nil
}

// ValidateClientConfig validates the client configuration
func (v *Validator) ValidateClientConfig(config *ClientConfig) error {
    if config.AppKey == "" {
        return NewError(ErrValidation, "app key cannot be empty", nil)
    }
    
    if config.AppSecret == "" {
        return NewError(ErrValidation, "app secret cannot be empty", nil)
    }
    
    if config.BaseURL == "" {
        return NewError(ErrValidation, "base URL cannot be empty", nil)
    }
    
    if config.Timeout < 0 {
        return NewError(ErrValidation, "timeout cannot be negative", nil)
    }
    
    if config.RetryCount < 0 {
        return NewError(ErrValidation, "retry count cannot be negative", nil)
    }
    
    return nil
}