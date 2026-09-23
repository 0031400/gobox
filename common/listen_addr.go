package common

import (
	"net"
	"strconv"
)

type ListenAddr struct {
	Ip   net.IP
	Port uint16
}

func (t *ListenAddr) ToString() string {
	return net.JoinHostPort(t.Ip.String(), strconv.Itoa(int(t.Port)))
}
