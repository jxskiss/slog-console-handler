package bslog

const ErrorKey = "error"

var errorFormatter = func(key string, err error) Attr {
	if err == nil {
		return Attr{}
	}
	return String(ErrorKey, err.Error())
}

func SetErrorFormatter(f func(key string, err error) Attr) {
	errorFormatter = f
}

func ErrorAttr(err error) Attr {
	return errorFormatter(ErrorKey, err)
}
