# slog-console-handler

Package `slog-console-handler` implements a zero-dependency slog.Handler
that writes colorized logs to console.

Its output format is friendly for human reading in console.
The output format can be configured using HandlerOptions, which is
a drop-in replacement for slog.HandlerOptions.

`go get github.com/jxskiss/slog-console-handler@latest`

## Usage

```go
// Use the default handler.
slog.SetDefault(slog.New(slogconsolehandler.Default))

// Or, use custom HandlerOptions.
slog.SetDefault(slog.New(slogconsolehandler.New(os.Stderr, &slogconsolehandler.HandlerOptions{
    // ...
})))
```
