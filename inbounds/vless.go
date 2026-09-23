package inbounds

import "net"

type VlessInbound struct {
	ip   net.IP
	port uint16
	UUID [16]byte
}

func (v *VlessInbound) Start() {

}
