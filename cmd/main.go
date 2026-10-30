package main

import (
	"coderchirag/folio-tracker/lib/init_mod"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	err := init_mod.InitApp()
	if err != nil {
		slog.Error("init-app failed", "err", err)
		os.Exit(1)
	}
}
