package listeners

import (
	"crypto/tls"
	"gobox/common"
	"gobox/connections"
	tlsUtil "gobox/tls_util"
	"net"
)

type TcpListener struct {
	Addr     common.ListenAddr
	Listener net.Listener
	Tls      tlsUtil.TlsServerConfig
}

func NewTcpListener(addr common.ListenAddr, tls tlsUtil.TlsServerConfig) *TcpListener {
	return &TcpListener{Addr: addr, Tls: tls}
}
func (t *TcpListener) Start() error {
	var err error
	if t.Tls.Enabled {
		cert, err := tls.LoadX509KeyPair(t.Tls.CertificatePath, t.Tls.KeyPath)
		if err != nil {
			return err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		t.Listener, err = tls.Listen("tcp", t.Addr.ToString(), config)
		return err
	}
	t.Listener, err = net.Listen("tcp", t.Addr.ToString())
	return err
}

func (t *TcpListener) Accept() (connections.Connection, error) {
	conn, err := t.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return connections.NewTcpConnection(conn), nil
}
