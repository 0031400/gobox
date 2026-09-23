package outbounds

import (
	"encoding/binary"
	"errors"
	"gobox/connections"
	"gobox/transports"
)

type VlessOutbound struct {
	UUID      [16]byte
	Transport transports.Transport
}

func NewVlessOutbound(uuid [16]byte, transport transports.Transport) *VlessOutbound {
	return &VlessOutbound{UUID: uuid, Transport: transport}

}
func (v *VlessOutbound) Connect(session OutSession) (connections.Connection, error) {
	conn, err := v.Transport.Connect()
	if err != nil {
		return nil, err
	}
	var reqData []byte
	reqData = append(reqData, 0)
	reqData = append(reqData, v.UUID[:]...)
	reqData = append(reqData, 0)
	reqData = append(reqData, 1)
	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, session.Target.Port)
	reqData = append(reqData, portBuf...)
	if session.Target.IsIp {
		ip := session.Target.Ip
		if ip.To4() != nil {
			reqData = append(reqData, 1)
			reqData = append(reqData, ip.To4()...)
		} else {
			reqData = append(reqData, 3)
			reqData = append(reqData, ip.To16()...)
		}
	} else {
		domain := session.Target.Domain
		reqData = append(reqData, 2, byte(len(domain)))
		reqData = append(reqData, []byte(domain)...)
	}
	reqData = append(reqData, session.FirstData...)
	err = conn.Write(reqData)
	if err != nil {
		return nil, err
	}
	res, err := conn.ReadExactly(2)
	if err != nil {
		return nil, err
	}
	if res[0] != 0 || res[1] != 0 {
		return nil, errors.New("error res")
	}
	return conn, nil
}
