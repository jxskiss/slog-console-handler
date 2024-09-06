package main

import (
	"log"
	"log/slog"

	"github.com/jxskiss/slog-console-handler"
	"github.com/jxskiss/slog-console-handler/examples/dolog"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	slog.SetDefault(slog.New(slogconsolehandler.Default))
	dolog.DoLogging()
}
