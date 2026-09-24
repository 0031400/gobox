package tlsUtil

type TlsClientConfig struct {
	Enabled    bool
	ServerName string
	Insecure   bool
}
type TlsServerConfig struct {
	Enabled         bool
	CertificatePath string
	KeyPath         string
}
