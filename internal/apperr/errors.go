package apperr

import "fmt"

type Code string

const (
	Invalid      Code = "invalid_request"
	Unauthorized Code = "unauthorized"
	Forbidden    Code = "forbidden"
	NotFound     Code = "not_found"
	Conflict     Code = "conflict"
	Unavailable  Code = "unavailable"
	Internal     Code = "internal"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
func (e *Error) Unwrap() error                { return e.Err }
func New(c Code, m string) *Error             { return &Error{Code: c, Message: m} }
func Wrap(c Code, m string, err error) *Error { return &Error{Code: c, Message: m, Err: err} }
