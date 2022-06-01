package errs

var (
	ParamErr         = NewInnerError("param error", 400)
	TableNotFoundErr = NewInnerError("table not found", 404)
	CacheDeleteErr   = NewInnerError("cache delete failed", 400)
)
