package sourcesupdate

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config map[string]map[string]Category

type Category struct {
	Type    string   `yaml:"type"`
	Sources []string `yaml:"sources"`
}

type Target struct {
	File    string
	Type    string
	Sources []string
}

func LoadTargets(configPath string) ([]Target, error) {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", configPath, err)
	}

	var targets []Target
	for groupName, categories := range config {
		for categoryName, category := range categories {
			targets = append(targets, Target{
				File:    filepath.Join("data", groupName, "pending", categoryName),
				Type:    category.Type,
				Sources: category.Sources,
			})
		}
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].File < targets[j].File
	})

	return targets, nil
}

var httpClient = &http.Client{Timeout: 120 * time.Second}

func fetchURL(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/plain,text/html,application/octet-stream,*/*")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

var (
	adguardDomainRE    = regexp.MustCompile(`^\|\|([a-zA-Z0-9][a-zA-Z0-9\-.]*\.[a-zA-Z]{2,})\^(\$|$)`)
	domainLooksValidRE = regexp.MustCompile(`^[a-z0-9*]([a-z0-9\-*]*[a-z0-9*])?(\.[a-z0-9*]([a-z0-9\-*]*[a-z0-9*])?)+$`)
	cidrLooksValidRE   = regexp.MustCompile(`^[0-9a-fA-F:.]+/[0-9]{1,3}$`)
	clashPayloadLineRE = regexp.MustCompile(`(?i)^-\s*(DOMAIN|DOMAIN-SUFFIX)\s*,\s*([^,\s]+)`)
)

func parseAdGuard(data []byte) map[string]bool {
	out := make(map[string]bool)
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "@@") || !strings.HasPrefix(line, "||") {
			continue
		}

		if m := adguardDomainRE.FindStringSubmatch(line); m != nil {
			domain := strings.ToLower(strings.TrimPrefix(m[1], "*."))
			out[domain] = true
		}
	}
	return out
}

func parsePlainDomainList(data []byte) map[string]bool {
	out := make(map[string]bool)
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for sc.Scan() {
		line := strings.ToLower(strings.TrimSpace(sc.Text()))
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") || strings.HasPrefix(line, ";") {
			continue
		}
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}

		line = strings.TrimPrefix(line, "*.")
		line = strings.TrimSuffix(line, ".")

		if domainLooksValidRE.MatchString(line) {
			out[line] = true
		}
	}
	return out
}

func parseCIDRPlain(data []byte) map[string]bool {
	out := make(map[string]bool)
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		for _, tok := range strings.Fields(line) {
			if _, _, err := net.ParseCIDR(tok); err == nil && cidrLooksValidRE.MatchString(tok) {
				out[tok] = true
			}
		}
	}
	return out
}

func parseClashPayload(data []byte) map[string]bool {
	out := make(map[string]bool)
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if m := clashPayloadLineRE.FindStringSubmatch(line); m != nil {
			domain := strings.ToLower(strings.TrimSuffix(m[2], "."))
			if domainLooksValidRE.MatchString(domain) {
				out[domain] = true
			}
		}
	}
	return out
}

func parseHosts(data []byte) map[string]bool {
	out := make(map[string]bool)
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if ip := net.ParseIP(fields[0]); ip != nil {
			domain := strings.ToLower(fields[1])
			if domain == "localhost" || domain == "localhost.localdomain" {
				continue
			}
			if domainLooksValidRE.MatchString(domain) {
				out[domain] = true
			}
		}
	}
	return out
}

func parseSourceData(url, sourceType string) (map[string]bool, error) {
	data, err := fetchURL(url)
	if err != nil {
		return nil, err
	}

	switch sourceType {
	case "adguard":
		return parseAdGuard(data), nil
	case "plain":
		return parsePlainDomainList(data), nil
	case "hosts":
		return parseHosts(data), nil
	case "cidr-plain":
		return parseCIDRPlain(data), nil
	case "clash-payload":
		return parseClashPayload(data), nil
	default:
		return nil, fmt.Errorf("unknown type: %q", sourceType)
	}
}

func normalizeForCompare(line string) string {
	line = strings.TrimSpace(line)
	for _, comment := range []string{"#", "//"} {
		if idx := strings.Index(line, " "+comment); idx > 0 {
			line = line[:idx]
		}
	}
	line = strings.TrimSpace(line)
	for _, prefix := range []string{"full:", "regexp:", "keyword:"} {
		line = strings.TrimPrefix(line, prefix)
	}
	line = strings.TrimPrefix(line, "+.")

	return strings.ToLower(strings.TrimSpace(line))
}

func loadExistingEntries(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return make(map[string]bool), nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make(map[string]bool)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[normalizeForCompare(line)] = true
	}

	return out, sc.Err()
}

func isPendingTarget(path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/pending/")
}

func fetchPromotedSiblingEntries(pendingPath string) (map[string]bool, error) {
	out := make(map[string]bool)
	if !isPendingTarget(pendingPath) {
		return out, nil
	}

	parentDir := filepath.Dir(filepath.Dir(pendingPath))
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		found, err := loadExistingEntries(filepath.Join(parentDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		for k := range found {
			out[k] = true
		}
	}

	return out, nil
}

func processTarget(t Target, dryRun bool) error {
	existing, err := loadExistingEntries(t.File)
	if err != nil {
		return fmt.Errorf("read %s: %w", t.File, err)
	}

	alreadyCategorized, err := fetchPromotedSiblingEntries(t.File)
	if err != nil {
		return fmt.Errorf("categorized for %s: %w", t.File, err)
	}

	if n := len(alreadyCategorized); n > 0 {
		fmt.Printf("  categorized: %d entries\n", n)
	}

	combined := make(map[string]bool)
	var sourceLabels []string

	for _, url := range t.Sources {
		fmt.Printf("  fetch %s\n", url)
		found, err := parseSourceData(url, t.Type)
		if err != nil {
			fmt.Printf("    skip: %v\n", err)
			continue
		}
		fmt.Printf("    got: %d\n", len(found))
		for k := range found {
			combined[k] = true
		}
		sourceLabels = append(sourceLabels, url)
	}

	var fresh []string
	for k := range combined {
		if existing[k] || alreadyCategorized[k] {
			continue
		}
		fresh = append(fresh, k)
	}
	sort.Strings(fresh)

	fmt.Printf("  stats: fetch %d, exist %d, categ %d, new %d\n",
		len(combined), len(existing), len(alreadyCategorized), len(fresh))

	if len(fresh) == 0 {
		return nil
	}

	if dryRun {
		limit := len(fresh)
		if limit > 15 {
			limit = 15
		}
		for _, d := range fresh[:limit] {
			fmt.Println("    +", d)
		}
		if len(fresh) > limit {
			fmt.Printf("    ... and %d more\n", len(fresh)-limit)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(t.File), 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", t.File, err)
	}

	f, err := os.OpenFile(t.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s: %w", t.File, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintf(w, "\n# === auto-update | %s | %s ===\n", strings.Join(sourceLabels, ", "), time.Now().Format("2006-01-02"))

	isIPTarget := strings.Contains(filepath.ToSlash(t.File), "/ips/")

	for _, d := range fresh {
		if isIPTarget {
			fmt.Fprintln(w, d)
		} else {
			fmt.Fprintln(w, "+."+d)
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Printf("  appended %d to %s\n", len(fresh), t.File)
	return nil
}

func Run(targets []Target, dryRun bool, only string) {
	for _, t := range targets {
		if only != "" && !strings.HasSuffix(t.File, only) {
			continue
		}
		fmt.Printf("\n=== %s ===\n", t.File)
		if err := processTarget(t, dryRun); err != nil {
			fmt.Printf("  error: %v\n", err)
		}
	}
}
