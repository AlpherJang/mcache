package errs

type InnerError struct {
	err  error
	code int
}

func (e InnerError) Code() int {
	return e.code
}

func (e InnerError) Error() error {
	return e.err
}

func (e InnerError) String() string {
	return e.err.Error()
}
