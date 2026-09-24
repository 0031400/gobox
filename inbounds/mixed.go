package inbounds

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gobox/common"
	"gobox/connections"
	"gobox/listeners"
	"log"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
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
func (m *MixedInbound) handleHttpNotConnect(conn connections.Connection, headerBody []byte, method string, path string, version string) {
	u, err := url.Parse(path)
	if err != nil {
		log.Println(err)
		return
	}
	var session InSession
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		log.Println(err)
		return
	}
	session.Target = common.TargetAddrFromHostPort(u.Hostname(), uint16(port))
	session.FirstData = append(session.FirstData, fmt.Appendf(nil, "%s %s %s\r\n", method, u.RequestURI(), version)...)
	session.FirstData = append(session.FirstData, headerBody...)
	m.channel <- session
}
func (m *MixedInbound) handleHttpConnect(conn connections.Connection, headerBody []byte, path string) {
	var session InSession
	session.FirstData = headerBody
	host, portStr, err := net.SplitHostPort(path)
	if err != nil {
		log.Println(err)
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Println(err)
		return
	}
	session.Target.Port = uint16(port)
	ip := net.ParseIP(host)
	if ip == nil {
		session.Target.IsIp = false
	} else {
		session.Target.IsIp = false
		session.Target.Ip = ip
	}
	session.Conn = conn
	m.channel <- session
}
func (m *MixedInbound) handleHttp(conn connections.Connection, firstByte byte) {
	buffer := []byte{firstByte}
	for !bytes.Contains(buffer, []byte("\r\n")) {
		data, err := conn.Read(4096)
		if err != nil {
			log.Println(err)
			return
		}
		buffer = append(buffer, data...)
	}

	idx := bytes.Index(buffer, []byte("\r\n"))
	line := buffer[:idx]
	p1 := bytes.IndexByte(line, ' ')
	if p1 < 0 {
		log.Println("not found http method")
		return
	}
	method := string(line[:p1])
	p2 := bytes.IndexByte(line[p1+1:], ' ')
	if p2 < 0 {
		log.Println("not found http url")
		return
	}
	p2 += p1 + 1
	url := string(line[p1+1 : p2])
	version := string(line[p2+1:])
	if strings.ToLower(method) == "connect" {
		m.handleHttpConnect(conn, buffer[idx+2:], url)
	} else {
		m.handleHttpNotConnect(conn, buffer[idx+2:], method, url, version)
	}
}
func (m *MixedInbound) handleSocks5(conn connections.Connection) {
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
	ver, err := conn.ReadExactly(1)
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
func (m *MixedInbound) handle(conn connections.Connection) {
	ver, err := conn.ReadExactly(1)
	if err != nil {
		log.Println(err)
		return
	}
	if ver[0] == 5 {
		m.handleSocks5(conn)
	} else {
		m.handleHttp(conn, ver[0])
	}
}
