package errs

import (
	"errors"
)

var (
	ParamErr = InnerError{err: errors.New("param error"), code: 400}
)
