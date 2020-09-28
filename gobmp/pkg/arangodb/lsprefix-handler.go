package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type lsPrefixArangoMessage struct {
	*message.LSPrefix
}

func (p *lsPrefixArangoMessage) MakeKey() string {
	return p.Prefix + "_" + strconv.Itoa(int(p.PrefixLen)) + "_" + p.IGPRouterID
}
