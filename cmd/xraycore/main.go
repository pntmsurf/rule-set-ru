package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/pntmsurf/rule-set-ru/internal/xraygeo"
)

func main() {
	root := flag.String("root", ".", "repo root with data folder")
	out := flag.String("out", "", "output folder for geosite.dat and geoip.dat, default root/release")
	flag.Parse()

	outDir := *out
	if outDir == "" {
		outDir = filepath.Join(*root, "release")
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("cannot create %s: %v", outDir, err)
	}

	domainsDir := filepath.Join(*root, "data", "domains")
	ipsDir := filepath.Join(*root, "data", "ips")

	geosite, err := xraygeo.BuildGeoSiteList(domainsDir)
	if err != nil {
		log.Fatalf("build geosite: %v", err)
	}
	geositePath := filepath.Join(outDir, "geosite.dat")
	if err := xraygeo.WriteGeoSite(geositePath, geosite); err != nil {
		log.Fatalf("write %s: %v", geositePath, err)
	}
	log.Printf("%s (%d categories)", geositePath, len(geosite.Entry))

	geoip, err := xraygeo.BuildGeoIPList(ipsDir)
	if err != nil {
		log.Fatalf("build geoip: %v", err)
	}
	geoipPath := filepath.Join(outDir, "geoip.dat")
	if err := xraygeo.WriteGeoIP(geoipPath, geoip); err != nil {
		log.Fatalf("write %s: %v", geoipPath, err)
	}
	log.Printf("%s (%d categories)", geoipPath, len(geoip.Entry))
}
