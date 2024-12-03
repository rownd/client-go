package testutils

// Client defines the interface needed for auth operations
type Client interface {
    GetBaseURL() string
    GetAppKey() string
} 