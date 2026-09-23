package inbounds

import (
	"encoding/binary"
	"fmt"
	"gobox/common"
	"gobox/connections"
	"gobox/listeners"
	"log"
	"slices"
)

type MixedInbound struct {
	Listener listeners.Listener
	channel  chan InSession
}

func NewMixedInbound(listener listeners.Listener) *MixedInbound {
	return &MixedInbound{Listener: listener, channel: make(chan InSession, 1000)}
}
func (m *MixedInbound) Start() error {
	err := m.Listener.Start()
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := m.Listener.Accept()
			if err != nil {
				log.Println(err)
				return
			}
			go m.handle(conn)
		}
	}()
	return nil
}

func (m *MixedInbound) Accept() InSession {
	return <-m.channel
}
func (m *MixedInbound) handle(conn connections.Connection) {
	ver, err := conn.ReadExactly(1)
	if err != nil {
		log.Println(err)
		return
	}
	if ver[0] != 5 {
		fmt.Println("version error")
		return
	}
	nMethod, err := conn.ReadExactly(1)
	if err != nil {
		log.Println(err)
		return
	}
	methods, err := conn.ReadExactly(int(nMethod[0]))
	if err != nil {
		log.Println(err)
		return
	}
	if !slices.Contains(methods, 0) {
		fmt.Println("auth error")
		return
	}
	err = conn.Write([]byte{5, 0})
	if err != nil {
		return
	}
	ver, err = conn.ReadExactly(1)
	if err != nil {
		return
	}
	if ver[0] != 5 {
		fmt.Println("version error")
		return
	}
	cmd, err := conn.ReadExactly(1)
	if err != nil {
		log.Println(err)
		return
	}
	if cmd[0] != 1 {
		fmt.Println("cmd error")
		return
	}
	_, err = conn.ReadExactly(1)
	if err != nil {
		log.Println(err)
		return
	}
	atyp, err := conn.ReadExactly(1)
	if err != nil {
		log.Println(err)
		return
	}
	var targetAddr common.TargetAddr
	switch atyp[0] {
	case 1:
		ip, err := conn.ReadExactly(4)
		if err != nil {
			log.Println(err)
			return
		}
		targetAddr.Ip = ip
		targetAddr.IsIp = true
	case 4:
		ip, err := conn.ReadExactly(16)
		if err != nil {
			log.Println(err)
			return
		}
		targetAddr.Ip = ip
		targetAddr.IsIp = true
	case 3:
		domainLen, err := conn.ReadExactly(1)
		if err != nil {
			log.Println(err)
			return
		}
		domain, err := conn.ReadExactly(int(domainLen[0]))
		if err != nil {
			log.Println(err)
			return
		}
		targetAddr.Domain = string(domain)
		targetAddr.IsIp = false
	default:
		fmt.Println("atyp error")
		return
	}
	portData, err := conn.ReadExactly(2)
	if err != nil {
		log.Println(err)
		return
	}
	targetAddr.Port = binary.BigEndian.Uint16(portData)
	err = conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return
	}
	var session InSession
	session.Conn = conn
	session.FirstData, err = conn.Read(4096)
	if err != nil {
		log.Println(err)
		return
	}
	session.Target = targetAddr
	m.channel <- session
}
