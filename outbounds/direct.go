package outbounds

import (
	"gobox/connections"
	tlsUtil "gobox/tls_util"
	"gobox/transports"
)

type DirectOutbound struct{}

func NewDirectOutbound() *DirectOutbound {
	return &DirectOutbound{}
}
func (v *DirectOutbound) Connect(session OutSession) (connections.Connection, error) {
	transport := transports.NewTcpTransport(session.Target, tlsUtil.TlsClientConfig{Enabled: false})
	conn, err := transport.Connect()
	if err != nil {
		return nil, err
	}
	err = conn.Write(session.FirstData)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
