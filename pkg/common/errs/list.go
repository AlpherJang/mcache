package errs

var (
	ParamErr         = NewInnerError("param error", 400)
	TableNotFoundErr = NewInnerError("table not found", 404)
	CacheDeleteErr   = NewInnerError("cache delete failed", 400)
	CacheNotFoundErr = NewInnerError("key not exist", 404)
	KeyExistsErr     = NewInnerError("key already exists", 400)
	UpdateCacheErr   = NewInnerError("update check rejected", 500)
)
