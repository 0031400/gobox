package router

import (
	"gobox/common"
	"net"
	"regexp"
	"slices"
	"strings"
)

type RouteRule struct {
	domain        []string
	domainSuffix  []string
	domainKeyword []string
	domainRegex   []regexp.Regexp
	ipCidr        []net.IPNet
	ruleSet       []RouteRule
	Outbound      string
}

func NewRouteRule(domain []string, domainSuffix []string, domainKeyword []string, domainRegex []string, ruleSet []RouteRule, ipCidr []string, outbound string) (*RouteRule, error) {
	var res []regexp.Regexp
	for _, item := range domainRegex {
		res = append(res, *regexp.MustCompile(item))
	}
	var cidrs []net.IPNet
	for _, item := range ipCidr {
		_, value, err := net.ParseCIDR(item)
		if err != nil {
			return nil, err
		}
		cidrs = append(cidrs, *value)
	}
	return &RouteRule{domain: domain, domainSuffix: domainSuffix, domainKeyword: domainKeyword, domainRegex: res, ruleSet: ruleSet, Outbound: outbound, ipCidr: cidrs}, nil
}

func (r *RouteRule) Match(addr common.TargetAddr) bool {
	for _, item := range r.ruleSet {
		if item.Match(addr) {
			return true
		}
	}
	if addr.IsIp {
		return r.matchIp(addr.Ip)
	}
	return r.matchDomain(addr.Domain)
}
func (r *RouteRule) matchIp(ip net.IP) bool {
	for _, item := range r.ipCidr {
		if item.Contains(ip) {
			return true
		}
	}
	return false
}
func (r *RouteRule) matchDomain(domain string) bool {
	if slices.Contains(r.domain, domain) {
		return true
	}
	for _, item := range r.domainSuffix {
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
	for _, item := range r.domain {
		if strings.Contains(domain, item) {
			return true
		}
	}
	for _, item := range r.domainRegex {
		if item.MatchString(domain) {
			return true
		}
	}
	return false
}
