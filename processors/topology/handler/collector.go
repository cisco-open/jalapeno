package handler

import (
	"github.com/cisco-ie/jalapeno/processors/topology/log"
	"github.com/cisco-ie/jalapeno/processors/topology/openbmp"
)

func collector(a *ArangoHandler, m *openbmp.Message) {
	if m.Action() != openbmp.ActionHeartbeat {
		log.Debugf("Got Collector %s [seq %v] action: %v.\n", m.GetUnsafe("admin_id"), m.GetUnsafe("sequence"), m.Action())
	}
}
