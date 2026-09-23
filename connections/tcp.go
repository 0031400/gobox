package connections

import (
	"net"
)

type TcpConnection struct {
	Conn   net.Conn
	Buffer []byte
}

func NewTcpConnection(conn net.Conn) *TcpConnection {
	return &TcpConnection{Conn: conn, Buffer: []byte{}}
}
func (t *TcpConnection) Read(n int) ([]byte, error) {
	if len(t.Buffer) == 0 {
		buf := make([]byte, n)
		nn, err := t.Conn.Read(buf)
		if err != nil {
			return nil, err
		}
		t.Buffer = buf[:nn]
	}
	if len(t.Buffer) > n {
		res := t.Buffer[:n]
		t.Buffer = t.Buffer[n:]
		return res, nil
	}
	res := t.Buffer
	t.Buffer = []byte{}
	return res, nil
}

func (t *TcpConnection) ReadExactly(n int) ([]byte, error) {
	for len(t.Buffer) < n {
		buf := make([]byte, n)
		nn, err := t.Conn.Read(buf)
		if err != nil {
			return nil, err
		}
		t.Buffer = append(t.Buffer, buf[:nn]...)
	}
	if len(t.Buffer) > n {
		res := t.Buffer[:n]
		t.Buffer = t.Buffer[n:]
		return res, nil
	}
	res := t.Buffer
	t.Buffer = []byte{}
	return res, nil
}

func (t *TcpConnection) Write(data []byte) error {
	_, err := t.Conn.Write(data)
	return err
}
func (t *TcpConnection) Close() {
	t.Conn.Close()
}
