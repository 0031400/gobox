package app

import (
	"errors"
	"gobox/config"
	"gobox/dns"
	dnsServers "gobox/dns/servers"
	"gobox/router"
	"net"
)

type Builder struct{ config config.AppConfig }

func NewBuilder(config config.AppConfig) *Builder {
	return &Builder{config: config}
}
func (b *Builder) buildDnsServer(cfg config.DnsServerConfig) (dnsServers.DnsServer, error) {
	if cfg.Type == "udp" {
		return dnsServers.NewUdpServer(net.UDPAddr{IP: net.ParseIP(cfg.Server), Port: int(cfg.ServerPort)}), nil
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
		rules = append(rules, *dns.NewDnsRule(cfg.Domain, cfg.DomainSuffix, cfg.DomainKeyword, cfg.DomainRegex, []dns.DnsRule{}, cfg.Server))
	}
	var listenAddr *net.UDPAddr
	if b.config.Dns.Listen != "" {
		listenAddr = &net.UDPAddr{IP: net.ParseIP(b.config.Dns.Listen), Port: int(b.config.Dns.ListenPort)}
	}
	return dns.NewDnsCenter(listenAddr, servers, rules, b.config.Dns.Final), nil
}
func (b *Builder) BuildRouter() (*router.Router, error) {
	var rules []router.RouteRule
	for _, rule := range b.config.Router.Rules {
		routeRule, err := router.NewRouteRule(rule.Domain, rule.DomainSuffix, rule.DomainKeyword, rule.DomainRegex, []router.RouteRule{}, rule.IpCidr, rule.Outbound)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *routeRule)
	}
	return router.NewRouter(b.config.Router.Final, rules), nil
}
