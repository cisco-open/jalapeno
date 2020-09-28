package arangodb

import (
	"github.com/sbezverk/gobmp/pkg/message"
)

type lsNodeArangoMessage struct {
	*message.LSNode
}

func (n *lsNodeArangoMessage) MakeKey() string {
	return n.RouterIP + "_" + n.PeerIP
}
