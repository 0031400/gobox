package listeners

import (
	"gobox/common"
	"gobox/connections"
	tlsUtil "gobox/tls_util"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WsListener struct {
	Addr    common.ListenAddr
	Path    string
	channel chan *websocket.Conn
	Tls     tlsUtil.TlsServerConfig
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	return true
}}

func NewWsListener(addr common.ListenAddr,
	path string,
	tls tlsUtil.TlsServerConfig) *WsListener {
	return &WsListener{Addr: addr, Path: path, Tls: tls, channel: make(chan *websocket.Conn, 1000)}
}
func (t *WsListener) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	t.channel <- conn
}
func (t *WsListener) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(t.Path, t.wsHandler)
	if t.Tls.Enabled {
		go http.ListenAndServeTLS(t.Addr.ToString(), t.Tls.CertificatePath, t.Tls.KeyPath, mux)
	} else {
		go http.ListenAndServe(t.Addr.ToString(), mux)
	}
	return nil
}

func (t *WsListener) Accept() (connections.Connection, error) {
	return connections.NewWsConnection(<-t.channel), nil
}
