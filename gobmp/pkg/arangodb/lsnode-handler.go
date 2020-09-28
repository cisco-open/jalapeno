package arangodb

import (
	"fmt"
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type lsNodeArangoMessage struct {
	*message.LSNode
}

func (n *lsNodeArangoMessage) MakeKey() string {

	protoIDStr := fmt.Sprintf("%d", n.ProtocolID)

	return protoIDStr + "_" + strconv.Itoa(int(n.DomainID)) + "_" + n.ISISAreaID + "_" + n.OSPFAreaID + "_" + n.IGPRouterID
}
