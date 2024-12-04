package config

import (
	"time"
)

// Config holds the internal configuration settings
type Config struct {
	APIVersion     string
	DefaultTimeout time.Duration
	MaxRetries     int
	UserAgent      string
	Endpoints      Endpoints
}

// Endpoints holds all API endpoint paths
type Endpoints struct {
	Auth  AuthEndpoints
	Users UsersEndpoints
}

type AuthEndpoints struct {
	Validate string
	Token    string
}

type UsersEndpoints struct {
	Get    string
	Update string
	Delete string
}

// NewConfig returns a new configuration with default values
func NewConfig() *Config {
	return &Config{
		APIVersion:     "v1",
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		UserAgent:      "rownd-go-sdk/1.0",
		Endpoints: Endpoints{
			Auth: AuthEndpoints{
				Validate: "/hub/auth/validate",
				Token:    "/hub/auth/token",
			},
			Users: UsersEndpoints{
				Get:    "/hub/users/%s",
				Update: "/hub/users/%s",
				Delete: "/hub/users/%s",
			},
		},
	}
}
