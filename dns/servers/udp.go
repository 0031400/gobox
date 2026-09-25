package servers

import "net"

type UdpServer struct {
	remote net.UDPAddr
}

func NewUdpServer(remote net.UDPAddr) *UdpServer {
	return &UdpServer{remote: remote}
}
func (u *UdpServer) Relay(data []byte) ([]byte, error) {
	conn, err := net.DialUDP("udp", nil, &u.remote)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], err
}
