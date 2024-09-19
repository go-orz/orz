package orz

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type ErrorCode int

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Wrapped error     `json:"-"`
}

func (e *Error) Error() string {
	if e.Wrapped == nil {
		return e.Message
	}
	return fmt.Errorf("%s: %w", e.Message, e.Wrapped).Error()
}

func (e *Error) Wrap(err error) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Wrapped: err,
	}
}

func (e *Error) Unwrap() error {
	return e.Wrapped
}

func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func ErrorFrom(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return NewError(ErrorCode(http.StatusInternalServerError), err.Error())
}
