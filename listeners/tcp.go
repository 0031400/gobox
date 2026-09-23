package listeners

import (
	"gobox/common"
	"gobox/connections"
	"net"
)

type TcpListener struct {
	Addr     common.ListenAddr
	Listener net.Listener
}

func NewTcpListener(addr common.ListenAddr) *TcpListener {
	return &TcpListener{Addr: addr}
}
func (t *TcpListener) Start() error {
	var err error
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
