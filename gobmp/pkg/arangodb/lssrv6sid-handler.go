package arangodb

import (
	"strconv"

	"github.com/sbezverk/gobmp/pkg/message"
)

type lsSRv6SIDArangoMessage struct {
	*message.LSSRv6SID
}

func (s *lsSRv6SIDArangoMessage) MakeKey() string {
	return strconv.Itoa(int(s.DomainID)) + "_" + s.IGPRouterID + "_" + s.SRv6SID
}
