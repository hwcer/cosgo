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
	ErrNotFound                    = NewHTTPError(http.StatusNotFound)
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
)

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message ...string) *HTTPError {
	he := &HTTPError{Code: code}
	if len(message) > 0 {
		he.Message = message[0]
	} else {
		he.Message = http.StatusText(code)
	}
	return he
}

// Error makes it compatible with `error` interface.
func (he *HTTPError) Error() string {
	return fmt.Sprintf("code:%d, message:%v", he.Code, he.Message)
}

func NewHTTPError500(message interface{}, debug bool) *HTTPError {
	var err string
	if m, ok := message.(error); ok && debug {
		err = m.Error()
	} else {
		err = http.StatusText(http.StatusInternalServerError)
	}
	return &HTTPError{Code: http.StatusInternalServerError, Message: err}
}
