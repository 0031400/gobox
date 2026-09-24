package servers

type DnsServer interface {
	Relay(data []byte) ([]byte, error)
}
