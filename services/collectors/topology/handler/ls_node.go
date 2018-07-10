package handler

import (
        "fmt"
        "strings"
        "wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
        "wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/openbmp"
)

func ls_node(a *ArangoHandler, m *openbmp.Message) {
        ls_sr := m.GetStr("ls_sr_capabilities")
        srgb_split := strings.Split(ls_sr, " ")
        srgb_start, srgb_range := srgb_split[2], srgb_split[1]
        // fmt.Println("Current message has srgb_start:", srgb_start)
        // fmt.Println("Current message has srgb_range:", srgb_range)

        name := m.GetStr("name")
        router_id := m.GetStr("router_id")
        combining_srgb := []string{srgb_start, srgb_range}
        combined_srgb := strings.Join(combining_srgb, ", ")
        // fmt.Println("Current message has srgb:", combined_srgb)

        // Parsing a Router from current LSNode OpenBMP message
        router_document := &database.Router{
                BGPID: router_id,
                SRGB: combined_srgb,
                Name: name,
        }
        if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("While upserting the current message's router document, encountered an error", err)
                return
        }
        fmt.Println("Successfully added Router:", router_id, "with SRGB:", combined_srgb, "and name:", name)

        // Parsing an Internal Transport Prefix from current LSNode OpenBMP message
        internal_transport_prefix_document := &database.InternalTransportPrefix{
                BGPID: router_id,
                SRGB: combined_srgb,
                Name: name,
        }
        if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("While upserting the current message's internal transport prefix document, encountered an error", err)
                return
        }
        fmt.Println("Successfully added Internal Transport Prefix:", router_id, "with SRGB:", combined_srgb, "and name:", name)

}
