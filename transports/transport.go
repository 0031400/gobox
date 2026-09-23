package transports

import "gobox/connections"

type Transport interface {
	Connect() (connections.Connection, error)
}
