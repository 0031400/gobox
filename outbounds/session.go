package outbounds

import "gobox/common"

type OutSession struct {
	Target    common.TargetAddr
	FirstData []byte
}
