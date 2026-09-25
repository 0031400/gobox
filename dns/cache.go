package dns

import (
	"log"
	"net"
	"sync"
	"time"
)

type DnsCacheCenter struct {
	data map[string]DnsCacheItem
	mu   sync.RWMutex
}
type DnsCacheItem struct {
	ips      []net.IP
	lastTime time.Time
}

const dnsCacheTtl = 3 * time.Second

func NewDnsCacheCenter() *DnsCacheCenter {
	return &DnsCacheCenter{data: make(map[string]DnsCacheItem)}
}
func (d *DnsCacheCenter) restore(domain string, ips []net.IP, dnsType int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	item, ok := d.data[domain]
	if !ok || (dnsType != 1 && dnsType != 28) {
		d.data[domain] = DnsCacheItem{ips: ips, lastTime: time.Now()}
	} else {
		for _, value := range item.ips {
			if dnsType == 1 && value.To16() != nil {
				ips = append(ips, value)
			}
			if dnsType == 28 && value.To4() != nil {
				ips = append(ips, value)
			}
		}
		d.data[domain] = DnsCacheItem{ips: ips, lastTime: time.Now()}
	}
	log.Printf("[dns] restore %s", domain)
}

func (d *DnsCacheCenter) lookup(domain string) []net.IP {
	d.mu.Lock()
	defer d.mu.Unlock()
	item, ok := d.data[domain]
	if !ok {
		return []net.IP{}
	}
	item.lastTime = time.Now()
	d.data[domain] = item
	log.Printf("[dns] lookup %s", domain)
	return item.ips
}
func (d *DnsCacheCenter) start() {
	go func() {
		timer := time.NewTicker(dnsCacheTtl)
		defer timer.Stop()
		for {
			<-timer.C
			d.mu.Lock()
			for domain, item := range d.data {
				if time.Since(item.lastTime) > dnsCacheTtl {
					delete(d.data, domain)
				}
			}
			d.mu.Unlock()
		}
	}()
}
