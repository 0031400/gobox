package common

import (
	"net"
	"strconv"
)

type TargetAddr struct {
	Ip     net.IP
	Domain string
	Port   uint16
	IsIp   bool
}

func (t *TargetAddr) ToString() string {
	if t.IsIp {
		return net.JoinHostPort(t.Ip.String(), strconv.Itoa(int(t.Port)))
	}
	return net.JoinHostPort(t.Domain, strconv.Itoa(int(t.Port)))
}
func TargetAddrFromHostPort(host string, port uint16) TargetAddr {
	ip := net.ParseIP(host)
	if ip != nil {
		return TargetAddr{Ip: ip, IsIp: true, Port: port}
	} else {
		return TargetAddr{Domain: host, IsIp: false, Port: port}
	}
}
