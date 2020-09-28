package arangodb

import (
	"fmt"
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type lsPrefixArangoMessage struct {
	*message.LSPrefix
}

func (p *lsPrefixArangoMessage) MakeKey() string {

	protoIDStr := fmt.Sprintf("%d", p.ProtocolID)

	return protoIDStr + "_" + strconv.Itoa(int(p.DomainID)) + "_" + strconv.Itoa(int(p.MTID)) + "_" + p.Prefix + "_" + strconv.Itoa(int(p.PrefixLen)) + "_" + p.IGPRouterID
}
