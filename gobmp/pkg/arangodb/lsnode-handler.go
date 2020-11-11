package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/base"
	"github.com/sbezverk/gobmp/pkg/message"
)

type lsNodeArangoMessage struct {
	*message.LSNode
}

func (n *lsNodeArangoMessage) MakeKey() string {
	area_id := "0"
	if (n.ProtocolID == base.OSPFv2 || n.ProtocolID == base.OSPFv3) {
		area_id = n.AreaID
	}
	// The LSNode Key uses ProtocolID, DomainID, and Area ID (if node is for OSPF protocol)
	// to create unique Keys for DB entries in multi-area / multi-topology scenarios
	return strconv.Itoa(int(n.ProtocolID)) + "_" + strconv.Itoa(int(n.DomainID)) + "_" + area_id + "_" + n.IGPRouterID
}
