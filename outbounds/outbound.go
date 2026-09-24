package outbounds

import "gobox/connections"

type Outbound interface {
	Connect(session OutSession) (connections.Connection, error)
}
