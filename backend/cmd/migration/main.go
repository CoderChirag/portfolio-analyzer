package migration

import (
	"coderchirag/folio-tracker/lib/config"
	"coderchirag/folio-tracker/lib/init_mod"
	"log/slog"
	"os"
)

func Main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})))

	cfg, err := config.Load()
	if err != nil {
		fatal("config parsing failed", err)
	}

	err = init_mod.InitApp(cfg)
	if err != nil {
		fatal("init-app failed", err)
	}
}

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
