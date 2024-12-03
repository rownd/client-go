package rowndmiddleware

import (
	"net/http"
	"strings"

	"github.com/rgthelen/rownd-go-sdk/pkg/rownd"
)

const (
	headerAuthentication string = "Authentication"
)

type (
	TokenExtractor func(r *http.Request) (string, error)
	ErrorHandler   func(w http.ResponseWriter, r *http.Request, err error)
)

type Handler struct {
	Validator      rownd.TokenValidator
	TokenExtractor TokenExtractor
	ErrorHandler   func(w http.ResponseWriter, r *http.Request, err error)
}

func NewHandler(validator rownd.TokenValidator, opts ...HandlerOption) (*Handler, error) {
	o := handlerOptions{
		errorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusForbidden)
		},
		tokenExtractor: func(r *http.Request) (string, error) {
			header := r.Header.Get(headerAuthentication)
			if header == "" {
				return "", nil
			}

			_, unverified, ok := strings.Cut(header, "Bearer ")
			if !ok {
				return "", nil
			}

			return unverified, nil
		},
	}
	for _, opt := range opts {
		opt.apply(&o)
	}
	if err := o.validate(); err != nil {
		return nil, err
	}

	h := &Handler{
		Validator:    validator,
		ErrorHandler: o.errorHandler,
	}

	return h, nil
}

type HandlerOption interface {
	apply(*handlerOptions)
}

type handlerOptions struct {
	errorHandler   func(w http.ResponseWriter, r *http.Request, err error)
	tokenExtractor TokenExtractor
}

func (o handlerOptions) validate() error {
	return nil
}

type errorHandlerOpt struct {
	fn func(w http.ResponseWriter, r *http.Request, err error)
}

func (o errorHandlerOpt) apply(opts *handlerOptions) {
	opts.errorHandler = o.fn
}

func WithErrorHandler(fn func(w http.ResponseWriter, r *http.Request, err error)) HandlerOption {
	return errorHandlerOpt{fn: fn}
}

type extractorOpt struct {
	fn TokenExtractor
}

func (o extractorOpt) apply(opts *handlerOptions) {
	opts.tokenExtractor = o.fn
}

func WithTokenExtractor(fn TokenExtractor) HandlerOption {
	return extractorOpt{fn: fn}
}
