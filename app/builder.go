package app

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"gobox/common"
	"gobox/config"
	"gobox/dns"
	dnsServers "gobox/dns/servers"
	"gobox/inbounds"
	"gobox/listeners"
	"gobox/outbounds"
	"gobox/router"
	tlsUtil "gobox/tls_util"
	"gobox/transports"
	"net"
	"os"
	"strings"
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
	Domain        config.Strings `json:"domain"`
	DomainKeyword config.Strings `json:"domain_keyword"`
	DomainSuffix  config.Strings `json:"domain_suffix"`
	DomainRegex   config.Strings `json:"domain_regex"`
	IpCidr        config.Strings `json:"ip_cidr"`
}
type RuleSetObject struct {
	Rules []RuleSetItem `json:"rules"`
}

func (b *Builder) LoadFileRuleSet(item config.RuleSetConfig) error {
	var ruleSet RuleSetObject
	file, err := os.Open(item.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&ruleSet)
	if err != nil {
		return err
	}
	var routeRules []router.RouteRule
	for _, value := range ruleSet.Rules {
		a, err := router.NewRouteRule(value.Domain, value.DomainSuffix, value.DomainKeyword, value.DomainRegex, []router.RouteRule{}, value.IpCidr, "")
		if err != nil {
			return err
		}
		routeRules = append(routeRules, *a)
	}
	b.ruleSets[item.Tag] = routeRules
	return nil
}
func (b *Builder) LoadRuleSet() error {
	for _, item := range b.config.Router.RuleSets {
		if item.Type == "local" {
			if item.Format == "source" {
				err := b.LoadFileRuleSet(item)
				if err != nil {
					return err
				}
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

func (b *Builder) buildListener(cfg config.InboundConfig) (listeners.Listener, error) {
	addr := common.ListenAddr{Ip: net.ParseIP(cfg.Listen), Port: cfg.ListenPort}
	tlsCfg := tlsUtil.TlsServerConfig{Enabled: cfg.TLs.Enabled, CertificatePath: cfg.TLs.CertificatePath, KeyPath: cfg.TLs.KeyPath}
	if cfg.Transport.Type == "tcp" || cfg.Transport.Type == "" {
		return listeners.NewTcpListener(addr, tlsCfg), nil
	} else if cfg.Transport.Type == "ws" {
		return listeners.NewWsListener(addr, cfg.Transport.Path, tlsCfg), nil
	} else {
		return nil, errors.New("unsupport listener type")
	}

}
func (b *Builder) BuildTransport(cfg config.OutboudConfig) (transports.Transport, error) {
	ip := net.ParseIP(cfg.Server)
	var addr common.TargetAddr
	addr.Port = cfg.ServerPort
	if ip == nil {
		addr.Domain = cfg.Server
		addr.IsIp = false
	} else {
		addr.Ip = ip
		addr.IsIp = true
	}
	tlsCfg := tlsUtil.TlsClientConfig{Enabled: cfg.TLs.Enabled, ServerName: cfg.TLs.ServerName, Insecure: cfg.TLs.Insecure}
	switch cfg.Transport.Type {
	case "tcp", "":
		return transports.NewTcpTransport(addr, tlsCfg), nil
	case "ws":
		return transports.NewWsTransport(addr, cfg.Transport.Host, cfg.Transport.Path, tlsCfg), nil
	default:
		return nil, errors.New("unsupport transport type")
	}
}
func (b *Builder) BuildOutbounds() (map[string]outbounds.Outbound, error) {
	items := make(map[string]outbounds.Outbound)
	for _, item := range b.config.Outbounds {
		transport, err := b.BuildTransport(item)
		if err != nil {
			return nil, err
		}
		switch item.Type {
		case "vless":
			var uuid [16]byte
			uuidStr := strings.ReplaceAll(item.User.Uuid, "-", "")
			if len(uuidStr) != 32 {
				return nil, errors.New("uuid len error")
			}
			_, err = hex.Decode(uuid[:], []byte(uuidStr))
			if err != nil {
				return nil, err
			}
			items[item.Tag] = outbounds.NewVlessOutbound(uuid, transport)
		case "direct":
			items[item.Tag] = outbounds.NewDirectOutbound()
		default:
			return nil, errors.New("unsupport inbound type")
		}
	}
	return items, nil
}
func (b *Builder) BuildInbounds() (map[string]inbounds.Inbound, error) {
	items := make(map[string]inbounds.Inbound)
	for _, item := range b.config.Inbounds {
		if item.Type == "mixed" {
			listener, err := b.buildListener(item)
			if err != nil {
				return nil, err
			}
			items[item.Tag] = inbounds.NewMixedInbound(listener)
		} else {
			return nil, errors.New("unsupport inbound type")
		}
	}
	return items, nil
}
