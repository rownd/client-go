package rownd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ClientOption ...
type ClientOption interface {
	apply(*clientOptions)
}

// clientOptions represents the configuration for the Rownd client
type clientOptions struct {
	appKey            string
	appSecret         string
	appID             string
	baseURL           string
	httpClient        *http.Client
	jwksCacheDuration time.Duration
	wkcCacheDuration  time.Duration
}

func (o clientOptions) validate() error {
	var errs []error

	if o.appKey == "" {
		errs = append(errs, errors.New("app key is required"))
	}
	if o.appSecret == "" {
		errs = append(errs, errors.New("app secret is required"))
	}
	if o.baseURL == "" {
		errs = append(errs, errors.New("base url is required"))
	}
	if _, err := url.Parse(o.baseURL); err != nil {
		errs = append(errs, fmt.Errorf("invalid base url: %w", err))
	}
	if o.httpClient == nil {
		errs = append(errs, errors.New("http client is required"))
	}
	if o.wkcCacheDuration < 0 {
		errs = append(errs, errors.New("well known config cache duration must be greater than zero"))
	}
	if o.jwksCacheDuration < 0 {
		errs = append(errs, errors.New("JSON Web Keys cache duration must be greater than zero"))
	}

	if len(errs) == 0 {
		return nil
	}

	return &MultiError{errors: errs}
}

type appKeyOpt string

func (o appKeyOpt) apply(opts *clientOptions) {
	opts.appKey = string(o)
}

// WithAppKey ...
func WithAppKey(appKey string) ClientOption {
	return appKeyOpt(appKey)
}

type appSecretOpt string

func (o appSecretOpt) apply(opts *clientOptions) {
	opts.appSecret = string(o)
}

// WithAppSecret ...
func WithAppSecret(secret string) ClientOption {
	return appSecretOpt(secret)
}

type appIDOpt string

func (o appIDOpt) apply(opts *clientOptions) {
	opts.appID = string(o)
}

// WithAppID ...
func WithAppID(appID string) ClientOption {
	return appIDOpt(appID)
}

type baseURLOpt string

func (o baseURLOpt) apply(opts *clientOptions) {
	opts.baseURL = string(o)
}

// WithBaseURL ...
func WithBaseURL(url string) ClientOption {
	return baseURLOpt(url)
}

type wkcCacheDurationOpt time.Duration

func (o wkcCacheDurationOpt) apply(opts *clientOptions) {
	opts.wkcCacheDuration = time.Duration(o)
}

// WithWKCCacheDuration ...
func WithWKCCacheDuration(d time.Duration) ClientOption {
	return wkcCacheDurationOpt(d)
}

type jwksCacheDurationOpt time.Duration

func (o jwksCacheDurationOpt) apply(opts *clientOptions) {
	opts.jwksCacheDuration = time.Duration(o)
}

// WithJWKsCacheDuration ...
func WithJWKsCacheDuration(d time.Duration) ClientOption {
	return jwksCacheDurationOpt(d)
}

// RequestOption ...
type RequestOption interface {
	apply(req *http.Request)
}

type requestOption struct {
	fn func(req *http.Request)
}

func (o requestOption) apply(req *http.Request) {
	o.fn(req)
}

// RequestWithHeader ...
func RequestWithHeader(key string, value string) RequestOption {
	return requestOption{
		fn: func(req *http.Request) {
			req.Header.Add(key, value)
		},
	}
}
