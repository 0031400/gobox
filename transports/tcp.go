package transports

import (
	"crypto/tls"
	"gobox/common"
	"gobox/connections"
	tlsUtil "gobox/tls_util"
	"net"
)

type TcpTransport struct {
	Addr common.TargetAddr
	Tls  tlsUtil.TlsClientConfig
}

func NewTcpTransport(addr common.TargetAddr,
	tls tlsUtil.TlsClientConfig) *TcpTransport {
	return &TcpTransport{Addr: addr, Tls: tls}
}
func (w *TcpTransport) Connect() (connections.Connection, error) {
	if w.Tls.Enabled {
		conn, err := tls.Dial("tcp", w.Addr.ToString(), &tls.Config{ServerName: w.Tls.ServerName, InsecureSkipVerify: w.Tls.Insecure})
		if err != nil {
			return nil, err
		}
		return connections.NewTcpConnection(conn), nil
	}
	conn, err := net.Dial("tcp", w.Addr.ToString())
	if err != nil {
		return nil, err
	}
	return connections.NewTcpConnection(conn), nil
}
