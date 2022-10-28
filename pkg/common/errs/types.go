package errs

import "errors"

type InnerError interface {
	Code() int
	Error() string
	GetError() error
}

func NewInnerError(err string, code int) InnerError {
	return &innerError{err: errors.New(err), code: code}
}

type innerError struct {
	err  error
	code int
}

func (e innerError) Code() int {
	return e.code
}

func (e innerError) GetError() error {
	return e.err
}

func (e innerError) Error() string {
	return e.err.Error()
}
