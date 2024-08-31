# slog-console-handler

Package slog-console-handler implements a zero-dependency slog.Handler
that writes colorized logs to console.

Its output format is friendly to human to read in console.
The output format can be customized using HandlerOptions which is a drop-in replacement for slog.HandlerOptinos.

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
