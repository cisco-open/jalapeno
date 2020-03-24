package handler

import (
	"github.com/cisco-ie/jalapeno/processors/topology/openbmp"
)

func unicast_prefix(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        // example: router_id := m.GetStr("router_ip")

        // Create and upsert any unicast_prefix documents in the future
}

