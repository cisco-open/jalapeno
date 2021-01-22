package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type lsLinkArangoMessage struct {
	*message.LSLink
}

func (l *lsLinkArangoMessage) MakeKey() string {
	var localID, remoteID string
	mtid := 0
	if l.MTID != nil {
		mtid = int(l.MTID.MTID)
	}
	if l.LocalLinkIP != "" && l.RemoteLinkIP != "" {
		localID = l.LocalLinkIP
		remoteID = l.RemoteLinkIP
	} else {
		localID = strconv.Itoa(int(l.LocalLinkID>>24)&0x000000ff) + "." +
			strconv.Itoa(int(l.LocalLinkID>>16)&0x000000ff) + "." +
			strconv.Itoa(int(l.LocalLinkID>>8)&0x000000ff) + "." +
			strconv.Itoa(int(l.LocalLinkID)&0x000000ff)
		remoteID = strconv.Itoa(int(l.RemoteLinkID>>24)&0x000000ff) + "." +
			strconv.Itoa(int(l.RemoteLinkID>>16)&0x000000ff) + "." +
			strconv.Itoa(int(l.RemoteLinkID>>8)&0x000000ff) + "." +
			strconv.Itoa(int(l.RemoteLinkID)&0x000000ff)
	}

	// The LSLink Key uses ProtocolID, DomainID, and Multi-Topology ID
	// to create unique Keys for DB entries in multi-area / multi-topology scenarios
	return strconv.Itoa(int(l.ProtocolID)) + "_" + strconv.Itoa(int(l.DomainID)) + "_" + strconv.Itoa(mtid) + "_" + l.AreaID + "_" + l.IGPRouterID + "_" + localID + "_" + l.RemoteIGPRouterID + "_" + remoteID
}
