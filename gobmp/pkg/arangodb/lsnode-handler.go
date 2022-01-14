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
	areaID := "0"
	if n.ProtocolID == base.OSPFv2 || n.ProtocolID == base.OSPFv3 {
		areaID = n.AreaID
	}
	// The LSNode Key uses ProtocolID, DomainID, and AreaID (if node is for OSPF protocol)
	// to create unique Keys for DB entries in multi-area / multi-topology scenarios
	return strconv.Itoa(int(n.ProtocolID)) + "_" + strconv.Itoa(int(n.DomainID)) + "_" + areaID + "_" + n.IGPRouterID
}
