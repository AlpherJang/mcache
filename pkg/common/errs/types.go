package errs

import (
	"errors"
)

type InnerError interface {
	Code() int
	String() string
	ToError() error
	Error() string
}

func NewInnerError(err string, code int) InnerError {
	return &innerError{err: errors.New(err), code: code}
}

type innerError struct {
	err  error
	code int
}

func (e innerError) Error() string {
	return e.String()
}

func (e innerError) Code() int {
	return e.code
}

func (e innerError) ToError() error {
	return e.err
}

func (e innerError) String() string {
	return e.err.Error()
}
