package inbounds

type Inbound interface {
	Start() error
	Accept() InSession
}
