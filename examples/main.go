package main

import (
	"log/slog"

	"github.com/jxskiss/slog-console-handler"
	"github.com/jxskiss/slog-console-handler/examples/dolog"
)

func main() {
	slog.SetDefault(slog.New(slogconsolehandler.Default))
	dolog.DoLogging()
}
