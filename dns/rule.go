package dns

import (
	"regexp"
	"slices"
	"strings"
)

type DnsRule struct {
	domain        []string
	domainSuffix  []string
	domainKeyword []string
	domainRegex   []*regexp.Regexp
	ruleSet       []DnsRule
	Server        string
}

func NewSubDnsRule(domain []string, domainSuffix []string, domainKeyword []string, domainRegex []*regexp.Regexp) *DnsRule {
	return &DnsRule{domain: domain, domainSuffix: domainSuffix, domainKeyword: domainKeyword, domainRegex: domainRegex, ruleSet: []DnsRule{}, Server: ""}
}
func NewDnsRule(domain []string, domainSuffix []string, domainKeyword []string, domainRegex []string, ruleSet []DnsRule, server string) *DnsRule {
	var res []*regexp.Regexp
	for _, item := range domainRegex {
		res = append(res, regexp.MustCompile(item))
	}
	return &DnsRule{domain: domain, domainSuffix: domainSuffix, domainKeyword: domainKeyword, domainRegex: res, ruleSet: ruleSet, Server: server}
}
func (d *DnsRule) Match(domain string) bool {
	domain = strings.TrimSuffix(domain, ".")
	if slices.Contains(d.domain, domain) {
		return true
	}
	for _, item := range d.domainSuffix {
		if domain == item {
			return true
		}
		if strings.HasPrefix(item, ".") {
			if strings.HasSuffix(domain, item) {
				return true
			}
		} else {
			if strings.HasSuffix(domain, "."+item) {
				return true
			}
		}
	}
	for _, item := range d.domain {
		if strings.Contains(domain, item) {
			return true
		}
	}
	for _, item := range d.domainRegex {
		if item.MatchString(domain) {
			return true
		}
	}
	for _, item := range d.ruleSet {
		if item.Match(domain) {
			return true
		}
	}
	return false
}
