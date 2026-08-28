package main

import (
	"flag"
	"log"

	"github.com/pntmsurf/rule-set-ru/internal/sourcesupdate"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "show changes but do not write files")
	only := flag.String("only", "", "update only file paths that end with this string")
	configPath := flag.String("config", "sources.yaml", "path to sources config file")
	flag.Parse()

	targets, err := sourcesupdate.LoadTargets(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	sourcesupdate.Run(targets, *dryRun, *only)
}
