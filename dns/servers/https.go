package servers

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
)

type HttpsServer struct {
	remote     net.TCPAddr
	path       string
	host       string
	serverName string
	insecure   bool
}

func NewHttpsServer(remote net.TCPAddr,
	path string,
	host string,
	serverName string,
	insecure bool) *HttpsServer {
	return &HttpsServer{remote: remote, path: path, host: host, serverName: serverName, insecure: insecure}
}
func (u *HttpsServer) Relay(data []byte) ([]byte, error) {
	conn, err := tls.Dial("tcp", u.remote.String(), &tls.Config{InsecureSkipVerify: u.insecure, ServerName: u.serverName})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s%s", u.host, u.path), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Host = u.host
	if err := req.Write(conn); err != nil {
		return nil, err
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
