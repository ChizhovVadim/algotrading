package main

import (
	"log/slog"

	"github.com/ChizhovVadim/algotrading/internal/cli"
)

func main() {
	var app = &cli.App{}
	app.AddCommand("test", testHandler)
	app.AddCommand("status", statusHandler)
	app.AddCommand("testdownload", testDownloadHandler)
	app.AddCommand("update", updateHandler)
	var err = app.Run()
	if err != nil {
		slog.Error("run failed",
			"error", err)
		return
	}
}
