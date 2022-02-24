package cosweb

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPError represents an error that occurred while handling a Request.
type HTTPError struct {
	Code    int         `json:"-"`
	Message interface{} `json:"message"`
}

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(http.StatusNotFound, "404 page not found")
	ErrUnauthorized                = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden                   = NewHTTPError(http.StatusForbidden)
	ErrMethodNotAllowed            = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrTooManyRequests             = NewHTTPError(http.StatusTooManyRequests)
	ErrBadRequest                  = NewHTTPError(http.StatusBadRequest)
	ErrBadGateway                  = NewHTTPError(http.StatusBadGateway)
	ErrInternalServerError         = NewHTTPError(http.StatusInternalServerError)
	ErrRequestTimeout              = NewHTTPError(http.StatusRequestTimeout)
	ErrServiceUnavailable          = NewHTTPError(http.StatusServiceUnavailable)
	ErrValidatorNotRegistered      = errors.New("validator not registered")
	ErrRendererNotRegistered       = errors.New("renderer not registered")
	ErrInvalidRedirectCode         = errors.New("invalid redirect status code")
	ErrCookieNotFound              = errors.New("cookie not found")
	ErrInvalidCertOrKeyType        = errors.New("invalid cert or key type, must be string or []byte")
	ErrArgsNotFound                = errors.New("args not found")
)

// Error makes it compatible with `error` interface.
func (he *HTTPError) Error() string {
	return he.String()
}

func (he *HTTPError) String() string {
	if he.Message != nil {
		return fmt.Sprintf("%v", he.Message)
	} else {
		code := he.Code
		if code == 0 {
			code = http.StatusOK
		}
		return http.StatusText(code)
	}
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message ...interface{}) *HTTPError {
	he := &HTTPError{Code: code}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

func NewHTTPError500(message interface{}) *HTTPError {
	return &HTTPError{Code: http.StatusInternalServerError, Message: message}
}
