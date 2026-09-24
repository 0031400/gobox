package app

import (
	"encoding/json"
	"errors"
	"gobox/config"
	"gobox/dns"
	dnsServers "gobox/dns/servers"
	"gobox/router"
	"net"
	"os"
)

type Builder struct {
	config   config.AppConfig
	ruleSets map[string][]router.RouteRule
}

func NewBuilder(config config.AppConfig) *Builder {
	return &Builder{config: config, ruleSets: make(map[string][]router.RouteRule)}
}
func (b *Builder) buildDnsServer(cfg config.DnsServerConfig) (dnsServers.DnsServer, error) {
	switch cfg.Type {
	case "udp":
		return dnsServers.NewUdpServer(net.UDPAddr{IP: net.ParseIP(cfg.Server), Port: int(cfg.ServerPort)}), nil
	case "https":
		return dnsServers.NewHttpsServer(net.TCPAddr{IP: net.ParseIP(cfg.Server), Port: int(cfg.ServerPort)}, cfg.Path, cfg.Host, cfg.ServerName, cfg.Insecure), nil
	}
	return nil, errors.New(("unsupport dns server type"))
}
func (b *Builder) BuildDnsCenter() (*dns.DnsCenter, error) {
	servers := make(map[string]dnsServers.DnsServer)
	for _, cfg := range b.config.Dns.Servers {
		server, err := b.buildDnsServer(cfg)
		if err != nil {
			return nil, err
		}
		servers[cfg.Tag] = server
	}
	var rules []dns.DnsRule
	for _, cfg := range b.config.Dns.Rules {
		var ruleSets []dns.DnsRule
		for _, item := range cfg.RuleSet {
			ruleSet, ok := b.ruleSets[item]
			if !ok {
				return nil, errors.New("rule set not found")
			}
			for _, value := range ruleSet {
				ruleSets = append(ruleSets, *dns.NewSubDnsRule(value.Domain, value.DomainSuffix, value.DomainKeyword, value.DomainRegex))
			}
		}
		rules = append(rules, *dns.NewDnsRule(cfg.Domain, cfg.DomainSuffix, cfg.DomainKeyword, cfg.DomainRegex, []dns.DnsRule{}, cfg.Server))
	}
	var listenAddr *net.UDPAddr
	if b.config.Dns.Listen != "" {
		listenAddr = &net.UDPAddr{IP: net.ParseIP(b.config.Dns.Listen), Port: int(b.config.Dns.ListenPort)}
	}
	return dns.NewDnsCenter(listenAddr, servers, rules, b.config.Dns.Final), nil
}

type RuleSetItem struct {
	Domain        []string `json:"domain"`
	DomainKeyword []string `json:"domain_keyword"`
	DomainSuffix  []string `json:"domain_suffix"`
	DomainRegex   []string `json:"domain_regex"`
	IpCidr        []string `json:"ip_cidr"`
}
type RuleSetObject struct {
	rules []RuleSetItem
}

func (b *Builder) LoadRuleSet() error {
	for _, item := range b.config.Router.RuleSets {
		if item.Type == "local" {
			if item.Format == "source" {
				var ruleSet RuleSetObject
				file, err := os.Open(item.Path)
				if err != nil {
					return err
				}
				defer file.Close()
				err = json.NewDecoder(file).Decode(&ruleSet)
				var routeRules []router.RouteRule
				for _, value := range ruleSet.rules {
					a, err := router.NewRouteRule(value.Domain, value.DomainSuffix, value.DomainKeyword, value.DomainRegex, []router.RouteRule{}, value.IpCidr, "")
					if err != nil {
						return err
					}
					routeRules = append(routeRules, *a)
				}
				b.ruleSets[item.Tag] = routeRules
			} else {
				return errors.New("unsupport rule set format")
			}
		} else {
			return errors.New("unsupport rule set type")
		}
	}
	return nil
}
func (b *Builder) BuildRouter() (*router.Router, error) {
	var rules []router.RouteRule
	for _, rule := range b.config.Router.Rules {
		var ruleSets []router.RouteRule
		for _, item := range rule.RuleSet {
			ruleSet, ok := b.ruleSets[item]
			if !ok {
				return nil, errors.New("rule set not found")
			}
			ruleSets = append(ruleSets, ruleSet...)
		}
		routeRule, err := router.NewRouteRule(rule.Domain, rule.DomainSuffix, rule.DomainKeyword, rule.DomainRegex, ruleSets, rule.IpCidr, rule.Outbound)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *routeRule)
	}
	return router.NewRouter(b.config.Router.Final, rules), nil
}
