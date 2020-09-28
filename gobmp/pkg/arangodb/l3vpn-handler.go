package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type l3VPNArangoMessage struct {
	*message.L3VPNPrefix
}

func (v *l3VPNArangoMessage) MakeKey() string {
	return v.VPNRD + "_" + v.Prefix + "_" + strconv.Itoa(int(v.PrefixLen))
}
