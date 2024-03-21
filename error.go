package betterslog

const ErrorKey = "error"

type errorValue struct{ error }

// Err returns an KindAny attribute for the supplied error,
// using "error" as key.
// HandlerOptions.ErrorFormatter can be used to format errors.
// If err is nil, it returns a zero Attr.
func Err(err error) Attr {
	return NamedErr(ErrorKey, err)
}

// NamedErr returns an KindAny attribute for the supplied error,
// using given key.
// HandlerOptions.ErrorFormatter can be used to format errors.
// If error is nil, it returns a zero Attr.
func NamedErr(key string, err error) Attr {
	if err == nil {
		return Attr{}
	}
	return Any(key, &errorValue{err})
}
