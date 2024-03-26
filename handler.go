package betterslog

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"time"
)

// Keys for "built-in" attributes.
const (
	TimeKey    = slog.TimeKey
	LevelKey   = slog.LevelKey
	MessageKey = slog.MessageKey
	SourceKey  = slog.SourceKey
)

type Handler = slog.Handler

// HandlerOptions are options for a handler, it extends slog.HandlerOptions.
// A zero HandlerOptions consists entirely of default values.
type HandlerOptions struct {
	AddSource   bool
	Level       Leveler
	ReplaceAttr func(groups []string, a Attr) Attr

	// AddLogger makes the handler to add a logger name attribute
	// to the output.
	AddLogger bool

	// CtxPerceptor specifies functions to detect level and get
	// additional attributes from context values.
	CtxPerceptor

	// ErrorFormatter sets a function to format errors.
	ErrorFormatter func(key string, err error) Attr

	// SourceFormatter sets a function to format the code location
	// to use instead of the default.
	// By default, SourceGroupFormatter is used.
	SourceFormatter func(s *Source) Attr

	// ConsoleOptions sets specific options for a ConsoleHandler.
	ConsoleOptions
}

// ConsoleOptions sets options for a ConsoleHandler.
type ConsoleOptions struct {
	// NoColor force disable color in console output.
	NoColor bool

	// TimeFormatter sets a function to format time in console output.
	// By default, TimeShortFormatter is used.
	TimeFormatter func(t time.Time) string
}

func (opts *HandlerOptions) newInnerHandlerOptions() *slog.HandlerOptions {
	inner := &slog.HandlerOptions{
		AddSource: opts.AddSource,
		Level:     opts.Level,
	}
	inner.ReplaceAttr = opts.getReplaceFunc()
	return inner
}

func (opts *HandlerOptions) getReplaceFunc() func(groups []string, a Attr) Attr {
	if opts.ReplaceAttr != nil ||
		opts.AddLogger ||
		opts.ErrorFormatter != nil ||
		opts.AddSource && opts.SourceFormatter != nil ||
		opts.ConsoleOptions.TimeFormatter != nil {
		return opts.replaceFunc
	}
	return nil
}

func (opts *HandlerOptions) replaceFunc(groups []string, a Attr) Attr {
	if len(groups) == 0 {
		switch a.Key {
		case TimeKey:
			if opts.ConsoleOptions.TimeFormatter != nil && a.Value.Kind() == KindTime {
				if t, ok := a.Value.Any().(time.Time); ok {
					return String(TimeKey, opts.ConsoleOptions.TimeFormatter(t))
				}
			}
		case SourceKey:
			if opts.AddSource && opts.SourceFormatter != nil && a.Value.Kind() == KindAny {
				if src, ok := a.Value.Any().(*Source); ok {
					return opts.SourceFormatter(src)
				}
			}
		case MessageKey:
			if opts.AddLogger {
				msg := a.Value.String()
				if strings.HasPrefix(msg, "@@logger=") {
					splitIdx := strings.IndexByte(msg, '\n')
					loggerName := msg[9:splitIdx]
					msg = msg[splitIdx+1:]
					a = Attr{Value: GroupValue(String(MessageKey, msg), String("logger", loggerName))}
				}
			}
		}
	}
	if opts.ErrorFormatter != nil && a.Value.Kind() == KindAny {
		if werr, ok := a.Value.Any().(*errorValue); ok {
			return opts.ErrorFormatter(a.Key, werr.error)
		}
	}
	if opts.ReplaceAttr != nil {
		a = opts.ReplaceAttr(groups, a)
	}
	return a
}

type CtxPerceptor struct {
	CheckLevel func(ctx context.Context) (level Level, change bool)
	CheckAttr  func(ctx context.Context) (attr Attr)
}

type JSONHandler struct {
	name  string
	opts  HandlerOptions
	inner *slog.JSONHandler
}

func NewJSONHandler(w io.Writer, opts *HandlerOptions) *JSONHandler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	innerOpts := opts.newInnerHandlerOptions()
	innerHandler := slog.NewJSONHandler(w, innerOpts)
	handler := &JSONHandler{
		opts:  *opts,
		inner: innerHandler,
	}
	return handler
}

func (h *JSONHandler) Enabled(ctx context.Context, level Level) bool {
	if ctxFunc := h.opts.CheckLevel; ctxFunc != nil {
		if ctxLevel, change := ctxFunc(ctx); change {
			return level >= ctxLevel
		}
	}
	if loggerName := h.getLoggerName(); loggerName != "" {
		if pll, ok := h.opts.Level.(*perLoggerLeveler); ok {
			return level >= pll.GetLoggerLevel(loggerName)
		}
	}
	return h.inner.Enabled(ctx, level)
}

func (h *JSONHandler) Handle(ctx context.Context, record Record) error {
	if h.opts.AddLogger && h.getLoggerName() != "" {
		record.Message = "@@logger=" + h.getLoggerName() + "\n" + record.Message
	}
	if ctxFunc := h.opts.CtxPerceptor.CheckAttr; ctxFunc != nil {
		ctxAttr := ctxFunc(ctx)
		if !ctxAttr.Equal(Attr{}) {
			record.AddAttrs(ctxAttr)
		}
	}
	return h.inner.Handle(ctx, record)
}

func (h *JSONHandler) clone() *JSONHandler {
	return &JSONHandler{
		name:  h.name,
		opts:  h.opts,
		inner: h.inner,
	}
}

func (h *JSONHandler) WithAttrs(attrs []Attr) Handler {
	clone := h.clone()
	clone.inner = h.inner.WithAttrs(attrs).(*slog.JSONHandler)
	return clone
}

func (h *JSONHandler) WithGroup(name string) Handler {
	clone := h.clone()
	clone.inner = h.inner.WithGroup(name).(*slog.JSONHandler)
	return clone
}

type TextHandler struct {
	name  string
	opts  HandlerOptions
	inner *slog.TextHandler
}

func NewTextHandler(w io.Writer, opts *HandlerOptions) *TextHandler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	innerOpts := opts.newInnerHandlerOptions()
	innerHandler := slog.NewTextHandler(w, innerOpts)
	handler := &TextHandler{
		opts:  *opts,
		inner: innerHandler,
	}
	return handler
}

func (h *TextHandler) Enabled(ctx context.Context, level Level) bool {
	if ctxFunc := h.opts.CheckLevel; ctxFunc != nil {
		if ctxLevel, change := ctxFunc(ctx); change {
			return level >= ctxLevel
		}
	}
	if loggerName := h.getLoggerName(); loggerName != "" {
		if pll, ok := h.opts.Level.(*perLoggerLeveler); ok {
			return level >= pll.GetLoggerLevel(loggerName)
		}
	}
	return h.inner.Enabled(ctx, level)
}

func (h *TextHandler) Handle(ctx context.Context, record Record) error {
	if h.opts.AddLogger && h.getLoggerName() != "" {
		record.Message = "@@logger=" + h.getLoggerName() + "\n" + record.Message
	}
	if ctxFunc := h.opts.CtxPerceptor.CheckAttr; ctxFunc != nil {
		ctxAttr := ctxFunc(ctx)
		if !ctxAttr.Equal(Attr{}) {
			record.AddAttrs(ctxAttr)
		}
	}
	return h.inner.Handle(ctx, record)
}

func (h *TextHandler) clone() *TextHandler {
	return &TextHandler{
		name:  h.name,
		opts:  h.opts,
		inner: h.inner,
	}
}

func (h *TextHandler) WithAttrs(attrs []Attr) Handler {
	clone := h.clone()
	clone.inner = h.inner.WithAttrs(attrs).(*slog.TextHandler)
	return clone
}

func (h *TextHandler) WithGroup(name string) Handler {
	clone := h.clone()
	clone.inner = h.inner.WithGroup(name).(*slog.TextHandler)
	return clone
}

type internalHandler interface {
	Handler
	cloneHandler() internalHandler
	getOptions() HandlerOptions
	getLoggerName() string
	setLoggerName(name string)
}

func (h *JSONHandler) cloneHandler() internalHandler { return h.clone() }
func (h *JSONHandler) getOptions() HandlerOptions    { return h.opts }
func (h *JSONHandler) getLoggerName() string         { return h.name }
func (h *JSONHandler) setLoggerName(name string)     { h.name = name }

func (h *TextHandler) cloneHandler() internalHandler { return h.clone() }
func (h *TextHandler) getOptions() HandlerOptions    { return h.opts }
func (h *TextHandler) getLoggerName() string         { return h.name }
func (h *TextHandler) setLoggerName(name string)     { h.name = name }
