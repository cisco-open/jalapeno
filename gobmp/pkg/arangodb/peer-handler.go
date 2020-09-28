package arangodb

import (
	"github.com/sbezverk/gobmp/pkg/message"
)

type peerStateChangeArangoMessage struct {
	*message.PeerStateChange
}

func (p *peerStateChangeArangoMessage) MakeKey() string {
	return p.RouterIP
}
