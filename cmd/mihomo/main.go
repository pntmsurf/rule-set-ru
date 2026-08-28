package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/pntmsurf/rule-set-ru/internal/dataset"
	"github.com/pntmsurf/rule-set-ru/internal/mihomobuild"
)

func main() {
	root := flag.String("root", ".", "repo root with data folder")
	out := flag.String("out", "", "output folder for .mrs files, default is root/release")
	mihomoBin := flag.String("mihomo-bin", "mihomo", "path to mihomo binary or name in PATH")
	flag.Parse()

	outDir := *out
	if outDir == "" {
		outDir = filepath.Join(*root, "release")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("cannot create %s: %v", outDir, err)
	}

	built := 0
	built += buildAll(filepath.Join(*root, "data", "domains"), outDir, mihomobuild.KindDomain, *mihomoBin)
	built += buildAll(filepath.Join(*root, "data", "ips"), outDir, mihomobuild.KindIPCIDR, *mihomoBin)
	log.Printf("built %d .mrs rule-sets in %s", built, outDir)
}

func buildAll(dir, outDir string, kind mihomobuild.Kind, mihomoBin string) int {
	files, err := dataset.List(dir)
	if err != nil {
		log.Fatalf("read %s: %v", dir, err)
	}

	n := 0
	for _, file := range files {
		has, err := dataset.HasEntries(file.Path)
		if err != nil {
			log.Fatalf("read category %s: %v", file.Path, err)
		}
		if !has {
			log.Printf("  skip empty %s", file.Path)
			continue
		}

		outPath := filepath.Join(outDir, file.Name+".mrs")
		if err := mihomobuild.Convert(mihomoBin, kind, file.Path, outPath); err != nil {
			log.Printf("  fail %s: %v", file.Name, err)
			continue
		}
		n++
	}
	return n
}
