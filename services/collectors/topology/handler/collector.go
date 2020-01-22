package handler

import (
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/log"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
)

func collector(a *ArangoHandler, m *openbmp.Message) {
	if m.Action() != openbmp.ActionHeartbeat {
		log.Debugf("Got Collector %s [seq %v] action: %v.\n", m.GetUnsafe("admin_id"), m.GetUnsafe("sequence"), m.Action())
	}
}
