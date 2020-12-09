package tripica

import (
	"fmt"
)

// Error represents any error coming out of service/tripica package.
type Error struct {
	Err error
}

// NewTriPicaError returns a new Error.
func NewTriPicaError(err error) *Error {
	return &Error{Err: err}
}

// Unwrap supports unwrapping of the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}

// Error returns the underlying error.
func (e *Error) Error() string {
	return fmt.Sprintf("service/tripica: %s", e.Err)
}
