package connections

import (
	"errors"

	"github.com/gorilla/websocket"
)

type WsConnection struct {
	Conn   *websocket.Conn
	Buffer []byte
}

func NewWsConnection(conn *websocket.Conn) *WsConnection {
	return &WsConnection{Conn: conn, Buffer: []byte{}}
}
func (w *WsConnection) Read(n int) ([]byte, error) {
	if len(w.Buffer) == 0 {
		messageType, message, err := w.Conn.ReadMessage()
		if err != nil {
			return nil, err
		}
		if messageType != websocket.BinaryMessage {
			return nil, errors.New("error ws message type")
		}
		w.Buffer = message
	}
	if len(w.Buffer) > n {
		res := w.Buffer[:n]
		w.Buffer = w.Buffer[n:]
		return res, nil
	}
	res := w.Buffer
	w.Buffer = []byte{}
	return res, nil
}

func (w *WsConnection) ReadExactly(n int) ([]byte, error) {
	for len(w.Buffer) < n {
		messageType, message, err := w.Conn.ReadMessage()
		if err != nil {
			return nil, err
		}
		if messageType != websocket.BinaryMessage {
			return nil, errors.New("error ws message type")
		}
		w.Buffer = append(w.Buffer, message...)
	}
	if len(w.Buffer) > n {
		res := w.Buffer[:n]
		w.Buffer = w.Buffer[n:]
		return res, nil
	}
	res := w.Buffer
	w.Buffer = []byte{}
	return res, nil
}

func (w *WsConnection) Write(data []byte) error {
	return w.Conn.WriteMessage(websocket.BinaryMessage, data)
}
func (w *WsConnection) Close() {
	w.Conn.Close()
}
