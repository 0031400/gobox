package transports

import (
	"crypto/tls"
	"fmt"
	"gobox/common"
	"gobox/connections"
	tlsUtil "gobox/tls_util"
	"net/http"

	"github.com/gorilla/websocket"
)

type WsTransport struct {
	Addr common.TargetAddr
	Host string
	Path string
	Tls  tlsUtil.TlsClientConfig
}

func NewWsTransport(addr common.TargetAddr,
	host string,
	path string,
	tls tlsUtil.TlsClientConfig) *WsTransport {
	return &WsTransport{Addr: addr, Host: host, Path: path, Tls: tls}
}
func (w *WsTransport) Connect() (connections.Connection, error) {
	dialer := websocket.Dialer{}
	if w.Tls.Enabled {
		dialer = websocket.Dialer{
			TLSClientConfig: &tls.Config{ServerName: w.Tls.ServerName, InsecureSkipVerify: w.Tls.Insecure},
		}
	}
	header := http.Header{}
	header.Set("Host", w.Host)
	protocol := "ws"
	if w.Tls.Enabled {
		protocol = "wss"
	}
	conn, _, err := dialer.Dial(fmt.Sprintf("%s://%s%s", protocol, w.Addr.ToString(), w.Path), header)
	if err != nil {
		return nil, err
	}
	return &connections.WsConnection{Conn: conn, Buffer: []byte{}}, nil
}
