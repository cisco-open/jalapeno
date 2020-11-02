package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type lsPrefixArangoMessage struct {
	*message.LSPrefix
}

func (p *lsPrefixArangoMessage) MakeKey() string {
	// The LSPrefix Key uses ProtocolID, DomainID, and Multi-Topology ID
	// to create unique Keys for DB entries in multi-area / multi-topology scenarios
	return strconv.Itoa(int(p.ProtocolID)) + "_" + strconv.Itoa(int(p.DomainID)) + "_" + strconv.Itoa(int(p.MTID)) + "_" + p.AreaID + "_" + strconv.Itoa(int(p.OSPFRouteType)) + "_" + p.Prefix + "_" + strconv.Itoa(int(p.PrefixLen)) + "_" + p.IGPRouterID
}
