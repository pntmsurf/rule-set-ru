package xraygeo

import (
	"net"
	"os"
	"strings"

	"github.com/v2fly/v2ray-core/v4/app/router"
	"google.golang.org/protobuf/proto"

	"github.com/pntmsurf/rule-set-ru/internal/dataset"
)

func BuildGeoSiteList(domainsDir string) (*router.GeoSiteList, error) {
	files, err := dataset.List(domainsDir)
	if err != nil {
		return nil, err
	}

	var list router.GeoSiteList
	for _, file := range files {
		lines, err := dataset.ReadLines(file.Path)
		if err != nil {
			return nil, err
		}

		geosite := &router.GeoSite{CountryCode: strings.ToUpper(file.Name)}
		for _, line := range lines {
			domainType := router.Domain_Domain
			switch {
			case strings.HasPrefix(line, "+."):
				domainType = router.Domain_Domain
				line = strings.TrimPrefix(line, "+.")
			case strings.HasPrefix(line, "domain:"):
				domainType = router.Domain_Domain
				line = strings.TrimPrefix(line, "domain:")
			case strings.HasPrefix(line, "full:"):
				domainType = router.Domain_Full
				line = strings.TrimPrefix(line, "full:")
			case strings.HasPrefix(line, "regexp:"):
				domainType = router.Domain_Regex
				line = strings.TrimPrefix(line, "regexp:")
			case strings.HasPrefix(line, "keyword:"):
				domainType = router.Domain_Plain
				line = strings.TrimPrefix(line, "keyword:")
			}

			if strings.Contains(line, "*") && domainType == router.Domain_Domain {
				domainType = router.Domain_Regex
				line = "^" + strings.ReplaceAll(strings.ReplaceAll(line, ".", `\.`), "*", ".*") + "$"
			}

			geosite.Domain = append(geosite.Domain, &router.Domain{
				Type:  domainType,
				Value: line,
			})
		}
		list.Entry = append(list.Entry, geosite)
	}
	return &list, nil
}

func BuildGeoIPList(ipsDir string) (*router.GeoIPList, error) {
	files, err := dataset.List(ipsDir)
	if err != nil {
		return nil, err
	}

	var list router.GeoIPList
	for _, file := range files {
		lines, err := dataset.ReadLines(file.Path)
		if err != nil {
			return nil, err
		}

		geoip := &router.GeoIP{CountryCode: strings.ToUpper(file.Name)}
		for _, line := range lines {
			ip, ipnet, err := net.ParseCIDR(line)
			if err != nil {
				ip = net.ParseIP(line)
				if ip == nil {
					continue
				}
				if ip.To4() != nil {
					ipnet = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
				} else {
					ipnet = &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
				}
			}

			ones, _ := ipnet.Mask.Size()
			ipBytes := ip.To4()
			if ipBytes == nil {
				ipBytes = ip.To16()
			}

			geoip.Cidr = append(geoip.Cidr, &router.CIDR{
				Ip:     ipBytes,
				Prefix: uint32(ones),
			})
		}
		list.Entry = append(list.Entry, geoip)
	}
	return &list, nil
}

func WriteGeoSite(path string, list *router.GeoSiteList) error {
	data, err := proto.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func WriteGeoIP(path string, list *router.GeoIPList) error {
	data, err := proto.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
