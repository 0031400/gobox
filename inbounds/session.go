package inbounds

import (
	"gobox/common"
	"gobox/connections"
)

type InSession struct {
	Conn      connections.Connection
	Target    common.TargetAddr
	FirstData []byte
}
