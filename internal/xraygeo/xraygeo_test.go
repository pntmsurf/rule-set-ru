package xraygeo

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/v2fly/v2ray-core/v4/app/router"
	"google.golang.org/protobuf/proto"
)

func TestBuildGeoSiteList(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantLen int
		wantErr bool
		check   func(t *testing.T, list *router.GeoSiteList)
	}{
		{
			name: "basic domain",
			files: map[string]string{
				"example": "example.com\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoSiteList) {
				if list.Entry[0].CountryCode != "EXAMPLE" {
					t.Fatalf("CountryCode = %s", list.Entry[0].CountryCode)
				}
				if len(list.Entry[0].Domain) != 1 {
					t.Fatalf("domains = %d", len(list.Entry[0].Domain))
				}
				d := list.Entry[0].Domain[0]
				if d.Type != router.Domain_Domain || d.Value != "example.com" {
					t.Fatalf("domain = %+v", d)
				}
			},
		},
		{
			name: "all prefixes",
			files: map[string]string{
				"test": "+.sub.example.com\ndomain:foo.com\nfull:bar.com\nregexp:^baz$\nkeyword:qux\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoSiteList) {
				domains := list.Entry[0].Domain
				if len(domains) != 5 {
					t.Fatalf("domains = %d", len(domains))
				}
				if domains[0].Type != router.Domain_Domain || domains[0].Value != "sub.example.com" {
					t.Fatalf("+. = %+v", domains[0])
				}
				if domains[1].Type != router.Domain_Domain || domains[1].Value != "foo.com" {
					t.Fatalf("domain: = %+v", domains[1])
				}
				if domains[2].Type != router.Domain_Full || domains[2].Value != "bar.com" {
					t.Fatalf("full: = %+v", domains[2])
				}
				if domains[3].Type != router.Domain_Regex || domains[3].Value != "^baz$" {
					t.Fatalf("regexp: = %+v", domains[3])
				}
				if domains[4].Type != router.Domain_Plain || domains[4].Value != "qux" {
					t.Fatalf("keyword: = %+v", domains[4])
				}
			},
		},
		{
			name: "wildcard to regex",
			files: map[string]string{
				"wild": "*.example.com\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoSiteList) {
				d := list.Entry[0].Domain[0]
				if d.Type != router.Domain_Regex {
					t.Fatalf("type = %v", d.Type)
				}
				if d.Value != `^.*\.example\.com$` {
					t.Fatalf("value = %s", d.Value)
				}
			},
		},
		{
			name: "wildcard with other prefix stays",
			files: map[string]string{
				"wild": "full:*.example.com\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoSiteList) {
				d := list.Entry[0].Domain[0]
				if d.Type != router.Domain_Full || d.Value != "*.example.com" {
					t.Fatalf("got = %+v", d)
				}
			},
		},
		{
			name: "multiple categories",
			files: map[string]string{
				"a": "a.com\n",
				"b": "b.com\n",
			},
			wantLen: 2,
			check: func(t *testing.T, list *router.GeoSiteList) {
				codes := map[string]bool{}
				for _, e := range list.Entry {
					codes[e.CountryCode] = true
				}
				if !codes["A"] || !codes["B"] {
					t.Fatalf("codes = %v", codes)
				}
			},
		},
		{
			name:    "empty dir",
			files:   map[string]string{},
			wantLen: 0,
			check:   func(t *testing.T, list *router.GeoSiteList) {},
		},
		{
			name: "empty lines and spaces",
			files: map[string]string{
				"empty": "\n  \nexample.com\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoSiteList) {
				if len(list.Entry[0].Domain) < 1 {
					t.Fatal("expected at least one domain")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				path := filepath.Join(dir, name+".txt")
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			list, err := BuildGeoSiteList(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(list.Entry) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(list.Entry), tt.wantLen)
			}
			tt.check(t, list)
		})
	}

	t.Run("list error", func(t *testing.T) {
		_, err := BuildGeoSiteList(filepath.Join(t.TempDir(), "no-such-dir"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestBuildGeoIPList(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantLen int
		wantErr bool
		check   func(t *testing.T, list *router.GeoIPList)
	}{
		{
			name: "cidr v4",
			files: map[string]string{
				"net": "1.2.3.0/24\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoIPList) {
				if list.Entry[0].CountryCode != "NET" {
					t.Fatalf("code = %s", list.Entry[0].CountryCode)
				}
				c := list.Entry[0].Cidr[0]
				if c.Prefix != 24 {
					t.Fatalf("prefix = %d", c.Prefix)
				}
				if net.IP(c.Ip).String() != "1.2.3.0" {
					t.Fatalf("ip = %v", c.Ip)
				}
			},
		},
		{
			name: "cidr v6",
			files: map[string]string{
				"net6": "2001:db8::/32\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoIPList) {
				c := list.Entry[0].Cidr[0]
				if c.Prefix != 32 {
					t.Fatalf("prefix = %d", c.Prefix)
				}
				if net.IP(c.Ip).String() != "2001:db8::" {
					t.Fatalf("ip = %v", c.Ip)
				}
			},
		},
		{
			name: "single ipv4",
			files: map[string]string{
				"host": "8.8.8.8\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoIPList) {
				c := list.Entry[0].Cidr[0]
				if c.Prefix != 32 {
					t.Fatalf("prefix = %d", c.Prefix)
				}
				if net.IP(c.Ip).String() != "8.8.8.8" {
					t.Fatalf("ip = %v", c.Ip)
				}
			},
		},
		{
			name: "single ipv6",
			files: map[string]string{
				"v6": "2001:db8::1\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoIPList) {
				c := list.Entry[0].Cidr[0]
				if c.Prefix != 128 {
					t.Fatalf("prefix = %d", c.Prefix)
				}
				if net.IP(c.Ip).String() != "2001:db8::1" {
					t.Fatalf("ip = %v", c.Ip)
				}
			},
		},
		{
			name: "skip invalid",
			files: map[string]string{
				"mixed": "1.1.1.1\nnot-an-ip\n2.2.2.0/24\n",
			},
			wantLen: 1,
			check: func(t *testing.T, list *router.GeoIPList) {
				if len(list.Entry[0].Cidr) != 2 {
					t.Fatalf("cidr count = %d", len(list.Entry[0].Cidr))
				}
			},
		},
		{
			name:    "empty dir",
			files:   map[string]string{},
			wantLen: 0,
			check:   func(t *testing.T, list *router.GeoIPList) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				path := filepath.Join(dir, name+".txt")
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			list, err := BuildGeoIPList(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(list.Entry) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(list.Entry), tt.wantLen)
			}
			tt.check(t, list)
		})
	}

	t.Run("list error", func(t *testing.T) {
		_, err := BuildGeoIPList(filepath.Join(t.TempDir(), "no-such-dir"))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestWriteGeoSite(t *testing.T) {
	list := &router.GeoSiteList{
		Entry: []*router.GeoSite{
			{
				CountryCode: "TEST",
				Domain: []*router.Domain{
					{Type: router.Domain_Domain, Value: "example.com"},
				},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "geosite.dat")
		if err := WriteGeoSite(path, list); err != nil {
			t.Fatal(err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var got router.GeoSiteList
		if err := proto.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if len(got.Entry) != 1 || got.Entry[0].CountryCode != "TEST" {
			t.Fatalf("got = %+v", got)
		}
	})

	t.Run("write error", func(t *testing.T) {
		dir := t.TempDir()
		if err := WriteGeoSite(dir, list); err == nil {
			t.Fatal("expected error writing to directory")
		}
	})
}

func TestWriteGeoIP(t *testing.T) {
	list := &router.GeoIPList{
		Entry: []*router.GeoIP{
			{
				CountryCode: "TEST",
				Cidr: []*router.CIDR{
					{Ip: net.ParseIP("1.2.3.0").To4(), Prefix: 24},
				},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "geoip.dat")
		if err := WriteGeoIP(path, list); err != nil {
			t.Fatal(err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var got router.GeoIPList
		if err := proto.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if len(got.Entry) != 1 || got.Entry[0].CountryCode != "TEST" {
			t.Fatalf("got = %+v", got)
		}
	})

	t.Run("write error", func(t *testing.T) {
		dir := t.TempDir()
		if err := WriteGeoIP(dir, list); err == nil {
			t.Fatal("expected error writing to directory")
		}
	})
}
