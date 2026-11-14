package main

import (
	"coderchirag/folio-tracker/cmd/api"
	"coderchirag/folio-tracker/cmd/jetgen"
	"coderchirag/folio-tracker/cmd/migration"
	"flag"
	"os"
)

func main() {
	app := flag.String("app", "", "app to start")
	out := flag.String("out", "./lib/db/repository", "output directory")
	flag.Parse()

	switch *app {
	case "migration":
		migration.Main()
	case "jetgen":
		jetgen.Main(*out)
	case "api":
		api.Main()
	default:
		flag.Usage()
		os.Exit(1)
	}
}
