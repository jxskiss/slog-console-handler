package betterslog

import (
	"log/slog"
	"sync"
	"time"
)

const (
	KindAny       = slog.KindAny
	KindBool      = slog.KindBool
	KindDuration  = slog.KindDuration
	KindFloat64   = slog.KindFloat64
	KindInt64     = slog.KindInt64
	KindString    = slog.KindString
	KindTime      = slog.KindTime
	KindUint64    = slog.KindUint64
	KindGroup     = slog.KindGroup
	KindLogValuer = slog.KindLogValuer
)

type (
	Attr      = slog.Attr
	Kind      = slog.Kind
	Value     = slog.Value
	LogValuer = slog.LogValuer
)

func String[T ~string](key string, v T) Attr    { return slog.String(key, string(v)) }
func Int64[T _SInt](key string, v T) Attr       { return slog.Int64(key, int64(v)) }
func Int[T _SInt](key string, v T) Attr         { return slog.Int(key, int(v)) }
func Uint64[T _UInt](key string, v T) Attr      { return slog.Uint64(key, uint64(v)) }
func Float64[T _Float](key string, v T) Attr    { return slog.Float64(key, float64(v)) }
func Bool(key string, v bool) Attr              { return slog.Bool(key, v) }
func Time(key string, v time.Time) Attr         { return slog.Time(key, v) }
func Duration(key string, v time.Duration) Attr { return slog.Duration(key, v) }
func Group(key string, args ...any) Attr        { return slog.Group(key, args...) }
func Any(key string, value any) Attr            { return slog.Any(key, value) }

func StringValue[T ~string](v T) Value    { return slog.StringValue(string(v)) }
func IntValue[T _SInt](v T) Value         { return slog.IntValue(int(v)) }
func Int64Value[T _SInt](v T) Value       { return slog.Int64Value(int64(v)) }
func Uint64Value[T _UInt](v T) Value      { return slog.Uint64Value(uint64(v)) }
func Float64Value[T _Float](v T) Value    { return slog.Float64Value(float64(v)) }
func BoolValue(v bool) Value              { return slog.BoolValue(v) }
func TimeValue(v time.Time) Value         { return slog.TimeValue(v) }
func DurationValue(v time.Duration) Value { return slog.DurationValue(v) }
func GroupValue(as ...Attr) Value         { return slog.GroupValue(as...) }
func AnyValue(v any) Value                { return slog.AnyValue(v) }

type _SInt interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int
}

type _UInt interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint | ~uintptr
}

type _Float interface {
	~float32 | ~float64 | _SInt | _UInt
}

type attrSlice []Attr

// pool of *[]Attr
var attrSlicePool = sync.Pool{
	New: func() any {
		as := make([]Attr, 0, 8)
		return (*attrSlice)(&as)
	},
}

func (as *attrSlice) Free() {
	*as = (*as)[:0]
	attrSlicePool.Put(as)
}
