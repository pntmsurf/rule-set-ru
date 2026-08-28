package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/pntmsurf/rule-set-ru/internal/dataset"
	"github.com/pntmsurf/rule-set-ru/internal/singboxrules"
)

func main() {
	root := flag.String("root", ".", "repo root with data folder")
	out := flag.String("out", "", "output folder for .srs files, default is root/release")
	singboxBin := flag.String("singbox-bin", "sing-box", "path to sing-box binary or name in PATH")
	flag.Parse()

	outDir := *out
	if outDir == "" {
		outDir = filepath.Join(*root, "release")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("cannot create %s: %v", outDir, err)
	}

	built := 0
	built += buildAll(filepath.Join(*root, "data", "domains"), outDir, false, *singboxBin)
	built += buildAll(filepath.Join(*root, "data", "ips"), outDir, true, *singboxBin)
	log.Printf("built %d .srs rule-sets in %s", built, outDir)
}

func buildAll(dir, outDir string, isIP bool, singboxBin string) int {
	files, err := dataset.List(dir)
	if err != nil {
		log.Fatalf("read %s: %v", dir, err)
	}

	n := 0
	for _, file := range files {
		srs, err := singboxrules.Build(file, isIP)
		if err != nil {
			log.Fatalf("read category %s: %v", file.Path, err)
		}

		suffix := ""
		if isIP {
			suffix = "-ip"
		}
		jsonPath := filepath.Join(outDir, file.Name+suffix+".json")
		srsPath := filepath.Join(outDir, file.Name+suffix+".srs")

		if err := singboxrules.Compile(singboxBin, srs, jsonPath, srsPath); err != nil {
			log.Printf("  fail %s: %v", file.Name, err)
			continue
		}
		n++
	}
	return n
}
