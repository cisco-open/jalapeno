package handler

import (
	"fmt"

	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/openbmp"
)

func router(a *ArangoHandler, m *openbmp.Message) {
        router_ip := m.GetStr("ip_addr")
        bgp_id := router_ip
        name := m.GetStr("name")
        
        // Parsing a Router document from current Peer OpenBMP message
        router_document := &database.Router{
		BGPID:    bgp_id,
		RouterIP: router_ip,
                Name:     name,
	}
	if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("While upserting the current message's router document, encountered an error")
	} else {
                fmt.Println("Successfully upserted router:", name, "with router_ip:", router_ip, "and bgp_id:", bgp_id)
        }

        // Parsing an Internal Transport Prefix document from current Peer OpenBMP message
	internal_transport_prefix_document := &database.InternalTransportPrefix{
		BGPID:    bgp_id,
		RouterIP: router_ip,
                Name:     name,
	}
	if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("While upserting the current message's internal transport prefix document, encountered an error")
	} else {
                fmt.Println("Successfully upserted internal transport prefix:", name, "with ip address:", router_ip, "and bgp_id:", bgp_id)
        }
}

