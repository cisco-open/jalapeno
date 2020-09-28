package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type unicastPrefixArangoMessage struct {
	*message.UnicastPrefix
}

func (u *unicastPrefixArangoMessage) MakeKey() string {
	return u.Prefix + "_" + strconv.Itoa(int(u.PrefixLen)) + "_" + u.PeerIP + "_" + u.Nexthop
}
