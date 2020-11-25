package errors

import (
	"fmt"
	"net/http"
)

// HTTPError represents an error that can occur while making HTTP calls with triPica.
type HTTPError struct {
	Err        error
	Body       string
	StatusCode int
}

// NewHTTPError returns a new HTTP error.
func NewHTTPError(err error, body []byte, statusCode int) *HTTPError {
	return &HTTPError{
		Err:        err,
		Body:       string(body),
		StatusCode: statusCode,
	}
}

// Error makes HTTPError implement the error interface.
func (e *HTTPError) Error() string {
	m := map[string]interface{}{}

	if e.Body != "" {
		m["body"] = e.Body
	}
	if e.StatusCode != 0 {
		m["status_code"] = e.StatusCode
	}
	if e.Err != nil {
		m["error"] = e.Err.Error()
	}

	return fmt.Sprintf("HTTP Error: %v", m)
}

// Unwrap supports unwrapping of the underlying error.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// Temporary determines whether the error in question is temporary.
func (e *HTTPError) Temporary() bool {
	return e.StatusCode != http.StatusBadRequest
}

// AuthorizationError represents an authorization error.
type AuthorizationError struct {
	Err error
}

// Error makes AuthorizationError implement the error interface.
func (e *AuthorizationError) Error() string {
	return e.Err.Error()
}

// Unwrap supports unwrapping of the underlying error.
func (e *AuthorizationError) Unwrap() error {
	return e.Err
}

// Temporary determines whether the error in question is temporary,
// meaning that the underlying operation can be retried.
func (e *AuthorizationError) Temporary() bool {
	return false
}

// NewHTTPRequestError returns a new error that occurs when making an HTTP request.
func NewHTTPRequestError(err error) error {
	fn := wthTrace(secondLevelTraceFunctionCall)
	if isHelperFunction(fn) {
		fn = wthTrace(thirdLevelTraceFunctionCall)
	}

	return &HTTPError{Err: fmt.Errorf("couldn't execute HTTP request %s: %w", fn, err)}
}

// NewParseError returns a new error that occurs when parsing an HTTP response body.
func NewParseError(err error, body []byte) error {
	fn := wthTrace(secondLevelTraceFunctionCall)
	if isHelperFunction(fn) {
		fn = wthTrace(thirdLevelTraceFunctionCall)
	}

	return &HTTPError{
		Err:  fmt.Errorf("couldn't parse response in %s: %w", fn, err),
		Body: string(body),
	}
}
