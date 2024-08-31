package slogconsolehandler

import (
	"context"
	"io"
	"log/slog"
)

// ConsoleHandler is a [slog.Handler] that writes log records to console,
// in a human-friendly format.
// It prints log message with color if the writer is a terminal that
// supports color.
type ConsoleHandler struct {
	inner *slog.TextHandler
}

// New creates a [ConsoleHandler] that writes to w,
// using the given options.
// If opts is nil, the default options are used.
// By default, color is enabled if w is a terminal and supports color.
func New(w io.Writer, opts *HandlerOptions) *ConsoleHandler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	enableColor := !opts.DisableColor && checkIsTerminal(w)
	cw := &consoleWriter{
		writer: w,
		color:  enableColor,
		buf:    make([]byte, 1024), // 1KB buffer
	}
	inner := slog.NewTextHandler(cw, &slog.HandlerOptions{
		AddSource:   opts.AddSource,
		Level:       opts.Level,
		ReplaceAttr: opts.ReplaceAttr,
	})
	return &ConsoleHandler{inner: inner}
}

func (h *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *ConsoleHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.inner.Handle(ctx, record)
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ConsoleHandler{
		inner: h.inner.WithAttrs(attrs).(*slog.TextHandler),
	}
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	return &ConsoleHandler{
		inner: h.inner.WithGroup(name).(*slog.TextHandler),
	}
}

// HandlerOptions are options for a [ConsoleHandler].
type HandlerOptions struct {
	// AddSource causes the handler to compute the source code position
	// of the log statement and add a SourceKey attribute to the output.
	AddSource bool

	// Level reports the minimum record level that will be logged.
	// The handler discards records with lower levels.
	// If Level is nil, the handler assumes LevelInfo.
	// The handler calls Level.Level for each record processed;
	// to adjust the minimum level dynamically, use a LevelVar.
	Level slog.Leveler

	// ReplaceAttr is called to rewrite each non-group attribute before it is logged.
	// The attribute's value has been resolved (see [Value.Resolve]).
	// If ReplaceAttr returns a zero Attr, the attribute is discarded.
	//
	// See slog.HandlerOptions for more details.
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr

	// DisableColor disables color output.
	DisableColor bool
}
