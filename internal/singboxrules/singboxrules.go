package singboxrules

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/pntmsurf/rule-set-ru/internal/dataset"
)

type SRS struct {
	Version int       `json:"version"`
	Rules   []SRSRule `json:"rules"`
}

type SRSRule struct {
	DomainSuffix  []string `json:"domain_suffix,omitempty"`
	DomainKeyword []string `json:"domain_keyword,omitempty"`
	DomainRegex   []string `json:"domain_regex,omitempty"`
	Domain        []string `json:"domain,omitempty"`
	IPCIDR        []string `json:"ip_cidr,omitempty"`
}

func RuleFromDomainLines(lines []string) SRSRule {
	var rule SRSRule
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "full:"):
			rule.Domain = append(rule.Domain, strings.TrimPrefix(line, "full:"))
		case strings.HasPrefix(line, "regexp:"):
			rule.DomainRegex = append(rule.DomainRegex, strings.TrimPrefix(line, "regexp:"))
		case strings.HasPrefix(line, "keyword:"):
			rule.DomainKeyword = append(rule.DomainKeyword, strings.TrimPrefix(line, "keyword:"))
		case strings.HasPrefix(line, "+."):
			bare := strings.TrimPrefix(line, "+.")
			// Если внутри префикса +. вдруг тоже затесалась звёздочка
			if strings.Contains(bare, "*") {
				regexStr := "^(.*\\.)?" + strings.ReplaceAll(strings.ReplaceAll(bare, ".", `\.`), "*", ".*") + "$"
				rule.DomainRegex = append(rule.DomainRegex, regexStr)
			} else {
				rule.Domain = append(rule.Domain, bare)
				rule.DomainSuffix = append(rule.DomainSuffix, "."+bare)
			}
		case strings.Contains(line, "*"):
			regexStr := "^" + strings.ReplaceAll(strings.ReplaceAll(line, ".", `\.`), "*", ".*") + "$"
			rule.DomainRegex = append(rule.DomainRegex, regexStr)
		default:
			rule.Domain = append(rule.Domain, line)
		}
	}
	return rule
}

func RuleFromIPLines(lines []string) SRSRule {
	return SRSRule{IPCIDR: lines}
}

func Build(file dataset.CategoryFile, isIP bool) (SRS, error) {
	lines, err := dataset.ReadLines(file.Path)
	if err != nil {
		return SRS{}, err
	}
	var rule SRSRule
	if isIP {
		rule = RuleFromIPLines(lines)
	} else {
		rule = RuleFromDomainLines(lines)
	}
	return SRS{Version: 1, Rules: []SRSRule{rule}}, nil
}

func Compile(singboxBin string, srs SRS, jsonPath, outSRSPath string) error {
	b, err := json.MarshalIndent(srs, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, b, 0644); err != nil {
		return err
	}
	defer os.Remove(jsonPath)

	cmd := exec.Command(singboxBin, "rule-set", "compile", jsonPath, "-o", outSRSPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
