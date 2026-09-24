package config

import (
	"encoding/json"
	"os"
)

type TlsClientConfig struct {
	Enabled    bool   `json:"enabled"`
	ServerName string `json:"server_name"`
	Insecure   bool   `json:"insecure"`
}
type TlsServerConfig struct {
	Enabled         bool   `json:"enabled"`
	CertificatePath string `json:"certificate_path"`
	KeyPath         string `json:"key_path"`
}
type TransportConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
	Host string `json:"host"`
}
type UserConfig struct {
	Name     string `json:"name"`
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type InboundConfig struct {
	Tag        string          `json:"tag"`
	Type       string          `json:"type"`
	Listen     string          `json:"listen"`
	ListenPort uint16          `json:"listen_port"`
	Users      []UserConfig    `json:"users"`
	TLs        TlsServerConfig `json:"tls"`
	Transport  TransportConfig `json:"transport"`
}
type OutboudConfig struct {
	Tag        string          `json:"tag"`
	Type       string          `json:"type"`
	Server     string          `json:"server"`
	ServerPort uint16          `json:"server_port"`
	TLs        TlsClientConfig `json:"tls"`
	User       UserConfig      `json:"user"`
	Transport  TransportConfig `json:"transport"`
}
type RuleSetConfig struct {
	Tag    string `json:"tag"`
	Type   string `json:"type"`
	Path   string `json:"path"`
	Format string `json:"format"`
}
type RuleConfig struct {
	Domain        []string `json:"domain"`
	DomainKeyword []string `json:"domain_keyword"`
	DomainSuffix  []string `json:"domain_suffix"`
	DomainRegex   []string `json:"domain_regex"`
	IpCidr        []string `json:"ip_cidr"`
	RuleSet       []string `json:"rule_set"`
	Outbound      string   `json:"outbound"`
}
type RouterConfig struct {
	Final    string          `json:"final"`
	RuleSets []RuleSetConfig `json:"rule_set"`
	Rules    []RuleConfig    `json:"rules"`
}
type DnsServerConfig struct {
	Tag        string `json:"tag"`
	Type       string `json:"type"`
	Server     string `json:"server"`
	ServerPort uint16 `json:"server_port"`
	Host       string `json:"host"`
	Path       string `json:"path"`
	ServerName string `json:"server_name"`
	Insecure   bool   `json:"insecure"`
}
type DnsRuleConfig struct {
	Domain        []string `json:"domain"`
	DomainKeyword []string `json:"domain_keyword"`
	DomainSuffix  []string `json:"domain_suffix"`
	DomainRegex   []string `json:"domain_regex"`
	RuleSet       []string `json:"rule_set"`
	Server        string   `json:"server"`
}
type DnsConfig struct {
	Listen     string            `json:"listen"`
	ListenPort uint16            `json:"listen_port"`
	Servers    []DnsServerConfig `json:"servers"`
	Rules      []DnsRuleConfig   `json:"rules"`
	Final      string            `json:"final"`
}
type AppConfig struct {
	Inbounds  []InboundConfig `json:"inbounds"`
	Outbounds []OutboudConfig `json:"outbounds"`
	Router    RouterConfig    `json:"router"`
	Dns       DnsConfig       `json:"dns"`
}

func LoadConfig(filePath string) (*AppConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var config AppConfig
	err = json.NewDecoder(file).Decode(&config)
	return &config, err
}
