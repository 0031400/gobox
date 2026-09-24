package dns

import (
	"errors"
	"fmt"
	"gobox/dns/servers"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
	mDns "github.com/miekg/dns"
)

var dnsCenter *DnsCenter

type DnsCenter struct {
	listenAddr *net.UDPAddr
	servers    map[string]servers.DnsServer
	rules      []DnsRule
	final      string
}

func NewDnsCenter(listenAddr *net.UDPAddr, servers map[string]servers.DnsServer, rules []DnsRule, final string) *DnsCenter {
	if dnsCenter != nil {
		return dnsCenter
	}
	dnsCenter = &DnsCenter{listenAddr: listenAddr, servers: servers, rules: rules, final: final}
	return dnsCenter
}
func Resolve(domain string) ([]net.IP, error) {
	if dnsCenter == nil {
		ips, err := net.LookupAddr(domain)
		if err != nil {
			return nil, err
		}
		var res []net.IP
		for _, ip := range ips {
			item := net.ParseIP(ip)
			if item != nil {
				res = append(res, item)
			}
		}
		return res, nil
	}
	return dnsCenter.Resolve(domain)
}
func (d *DnsCenter) Start() {
	if d.listenAddr == nil {
		return
	}
	go func() {
		conn, err := net.ListenUDP("udp", d.listenAddr)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()
		log.Printf("[dns] listen on %s\n", d.listenAddr)
		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := conn.ReadFromUDP(buf)
			log.Printf("[dns] connected: %s\n", clientAddr)
			if err != nil {
				fmt.Println(err)
				continue
			}
			go func(clientAddr net.UDPAddr, data []byte) {
				data, err := d.Relay(data)
				if err != nil {
					fmt.Println(err)
					return
				}
				_, err = conn.WriteToUDP(data, &clientAddr)
				if err != nil {
					fmt.Println(err)
				}
			}(*clientAddr, buf[:n])
		}
	}()
}
func (d *DnsCenter) Relay(data []byte) ([]byte, error) {
	msg := new(mDns.Msg)
	err := msg.Unpack(data)
	if err != nil {
		return nil, err
	}
	var domain string
	if len(msg.Question) == 0 {
		domain = ""
	} else {
		domain = msg.Question[0].Name
	}
	if domain != "" {
		log.Printf("[dns] relay: <- %s %d", domain, msg.Question[0].Qtype)
	}
	serverTag := d.findServerTag(domain)
	server, ok := d.servers[serverTag]
	if !ok {
		return nil, errors.New("dns server not found")
	}
	data, err = server.Relay(data)
	err = msg.Unpack(data)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, answer := range msg.Answer {
		if a, ok := answer.(*dns.A); ok {
			res = append(res, a.A.String())
		}
		if a, ok := answer.(*dns.AAAA); ok {
			res = append(res, a.AAAA.String())
		}
		if a, ok := answer.(*dns.CNAME); ok {
			res = append(res, a.Target)
		}
	}
	log.Printf("[dns] relay: %s -> %s -> %s", domain, serverTag, strings.Join(res, ","))
	return data, nil
}
func (d *DnsCenter) Resolve(domain string) ([]net.IP, error) {
	var ips, ips4, ips6 []net.IP
	var errOut error
	channel := make(chan struct{}, 2)
	go func() {
		var err error
		ips4, err = d.resolveOne(domain, false)
		if err != nil {
			errOut = err
		}
		channel <- struct{}{}
	}()
	go func() {
		var err error
		ips6, err = d.resolveOne(domain, true)
		if err != nil {
			errOut = err
		}
		channel <- struct{}{}
	}()
	<-channel
	<-channel
	ips = append(ips, ips4...)
	ips = append(ips, ips6...)
	if len(ips) == 0 && errOut != nil {
		return nil, errOut
	}
	return ips, nil
}
func (d *DnsCenter) resolveOne(domain string, v6 bool) ([]net.IP, error) {
	msg := new(mDns.Msg)
	if v6 {
		msg.SetQuestion(domain, dns.TypeAAAA)
	} else {
		msg.SetQuestion(domain, dns.TypeA)
	}
	msg.RecursionDesired = true
	data, err := msg.Pack()
	if err != nil {
		return nil, err
	}
	serverTag := d.findServerTag(domain)
	server, ok := d.servers[serverTag]
	if !ok {
		return nil, errors.New("dns server not found")
	}
	data, err = server.Relay(data)
	if err != nil {
		return nil, err
	}
	err = msg.Unpack(data)
	if err != nil {
		return nil, err
	}
	var ips []net.IP
	for _, answer := range msg.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, a.A)
		}
		if a, ok := answer.(*dns.AAAA); ok {
			ips = append(ips, a.AAAA)
		}
	}
	return ips, err
}
func (d *DnsCenter) findServerTag(domain string) string {
	serverTag := d.final
	for _, rule := range d.rules {
		if rule.Match(domain) {
			serverTag = rule.Server
			break
		}
	}
	return serverTag
}
