package listeners

import "gobox/connections"

type Listener interface {
	Start() error
	Accept() (connections.Connection, error)
}
