package rownd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"path"

	"github.com/patrickmn/go-cache"
)

type Sort string

const (
	SortAsc  Sort = "asc"
	SortDesc Sort = "desc"
)

const (
	headerRowndAppKey    string = "X-ROWND-APP-KEY"
	headerRowndAppSecret string = "X-ROWND-APP-SECRET"

	defaultBaseURL              string        = "https://api.rownd.io"
	defaultHTTPTimeout          time.Duration = 30 * time.Second
	defaultCacheTTL             time.Duration = 5 * time.Minute
	defaultCacheCleanupInterval time.Duration = 10 * time.Minute

	cacheKeyWKC  string = "wkc"
	cacheKeyJWKS string = "jwks"

	defaultWKCCacheDuration  time.Duration = 1 * time.Hour
	defaultJWKsCacheDuration time.Duration = 1 * time.Hour

	defaultJWKSPath = "/hub/auth/keys"
)

// ClientConfig contains the configuration for creating a new Rownd client
type ClientConfig struct {
	AppKey    string
	AppSecret string
	AppID     string
	BaseURL   string
	Timeout   time.Duration
}

// Client ...
type Client struct {
	appID          string
	appKey         string
	appSecret      string
	baseURL        string
	httpClient     *http.Client
	httpClientOpts []RequestOption

	// cache and cache timeouts
	cache             *cache.Cache
	wkcCacheDuration  time.Duration
	jwksCacheDuration time.Duration

	// client implementations
	Tokens       *tokenValidator
	Users        *userClient
	UserFields   *userFieldClient
	Groups       *groupClient
	GroupInvites *groupInviteClient
	GroupMembers *groupMemberClient
	MagicLinks   *magicLinkClient
	logger       *log.Logger
}

// NewClient creates a new Rownd client instance.
func NewClient(opts ...ClientOption) (*Client, error) {
	// build default set of options.
	o := clientOptions{
		appKey:            os.Getenv("ROWND_APP_KEY"),
		appSecret:         os.Getenv("ROWND_APP_SECRET"),
		appID:             os.Getenv("ROWND_APP_ID"),
		baseURL:           defaultBaseURL,
		httpClient:        &http.Client{Timeout: defaultHTTPTimeout},
		wkcCacheDuration:  defaultWKCCacheDuration,
		jwksCacheDuration: defaultJWKsCacheDuration,
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	if err := o.validate(); err != nil {
		return nil, err
	}

	// build client with validated options
	c := &Client{
		appKey:     o.appKey,
		appSecret:  o.appSecret,
		appID:      o.appID,
		baseURL:    o.baseURL,
		httpClient: o.httpClient,
		httpClientOpts: []RequestOption{
			RequestWithHeader(headerRowndAppKey, o.appKey),
			RequestWithHeader(headerRowndAppSecret, o.appSecret),
		},
		cache:             cache.New(defaultCacheTTL, defaultCacheCleanupInterval),
		wkcCacheDuration:  defaultWKCCacheDuration,
		jwksCacheDuration: defaultJWKsCacheDuration,
		logger:            log.New(os.Stdout, "[rownd] ", log.LstdFlags),
	}

	// build client implementations
	c.Tokens = &tokenValidator{c}
	c.Users = &userClient{c}
	c.UserFields = &userFieldClient{c}
	c.Groups = &groupClient{c}
	c.GroupInvites = &groupInviteClient{c}
	c.GroupMembers = &groupMemberClient{
		Client: c,
		logger: c.logger,
	}
	c.MagicLinks = &magicLinkClient{c}

	return c, nil
}

// rowndURL ...
func (c *Client) rowndURL(parts ...string) (*url.URL, error) {
	baseURL := c.baseURL
	if parts[0] == defaultJWKSPath {
		// For JWKS, use the base domain without /v1
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}
	
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	pathParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			pathParts = append(pathParts, part)
		}
	}

	endpoint.Path = path.Join(endpoint.Path, path.Join(pathParts...))
	return endpoint, nil
}

// request performs an HTTP request and unmarshals the response into v.
func (c *Client) request(ctx context.Context, method, url string, body, v interface{}, opts ...RequestOption) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("failed to marshal request payload: %w", err)
		}
	}

	// build HTTP request from arguments
	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Apply request options
	for _, opt := range opts {
		opt.apply(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx responses and handle them
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return handleErrorResponse(resp)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// For DELETE requests or empty responses, return nil
	if method == http.MethodDelete || len(respBody) == 0 {
		return nil
	}

	// Only try to unmarshal if we have a response target
	if v != nil {
		if err := json.Unmarshal(respBody, v); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}

// JWK represents a JSON Web Key.
type JWK struct {
	Alg string `json:"alg"`
	KTY string `json:"kty"`
	Use string `json:"use"`
	KID string `json:"kid"`
	CRV string `json:"crv"`
	X   string `json:"x"`
}

// JWKs represents a set of JSON Web Keys.
type JWKs struct {
	Keys []JWK `json:"keys"`
}

// Contains attempts to find the JSON Web Key that matches the supplied key id.
func (jwks JWKs) Contains(kid string) (JWK, bool) {
	if len(jwks.Keys) == 0 {
		return JWK{}, false
	}

	for _, key := range jwks.Keys {
		if key.KID == kid {
			return key, true
		}
	}

	return JWK{}, false
}

// fetchJWKS fetches the JSON Web Key Set directly
func (c *Client) fetchJWKS(ctx context.Context) (*JWKs, error) {
	cached, found := c.cache.Get(cacheKeyJWKS)
	if v, ok := cached.(*JWKs); found && ok {
		return v, nil
	}

	endpoint, err := c.rowndURL(defaultJWKSPath)
	if err != nil {
		return nil, err
	}

	var response *JWKs
	if err := c.request(ctx, http.MethodGet, endpoint.String(), nil, &response); err != nil {
		return nil, NewError(ErrAPI, "failed to fetch JWKS", err)
	}

	c.cache.Set(cacheKeyJWKS, response, c.jwksCacheDuration)
	return response, nil
}

// ToPointer is a generic function that returns a pointer of the supplied value.
func ToPointer[T any](value T) *T {
	return &value
}

// ToValue is a generic function that safely dereferences a pointer.
func ToValue[T any](value *T) T {
	var zero T
	if value == nil {
		return zero
	}
	return *value
}

// GetBaseURL returns the base URL for the client
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetAppKey returns the app key for the client
func (c *Client) GetAppKey() string {
	return c.appKey
}


