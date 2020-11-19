package arangodb

import (
	"net"
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type srPolicyArangoMessage struct {
	*message.SRPolicy
}

func (srp *srPolicyArangoMessage) MakeKey() string {
	var ep string
	if srp.IsIPv4 {
		ep = net.IP(srp.Endpoint).To4().String()
	} else {
		ep = net.IP(srp.Endpoint).To16().String()
	}

	return ep + "_" + srp.RouterIP + "_" + strconv.Itoa(int(srp.Distinguisher)) + "_" + strconv.Itoa(int(srp.Color))
}
