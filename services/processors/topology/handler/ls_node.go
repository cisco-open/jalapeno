package handler

import (
        "fmt"
        "strings"
	"strconv"
        "wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/database"
        "wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
)

func ls_node(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        router_ip     := m.GetStr("router_id")
        name          := m.GetStr("name")
        ls_sr         := m.GetStr("ls_sr_capabilities")
	igp_router_id := m.GetStr("igp_router_id")
 	bgp_id        := router_ip

        if ls_sr == "" {
                fmt.Println("No ls_sr_capabilities value available, skipping this OpenBMP message")
                return
        }
	srgb := parse_srgb(ls_sr)

	sr_node_sid := ""
	sr_prefix_sid := ""
        sr_beginning_label := parse_sr_beginning_label(srgb)
        sid_index := a.db.GetSIDIndex(bgp_id)
        if(sid_index != "") {
                sr_node_sid = calculate_sid(sr_beginning_label, sid_index)
                sr_prefix_sid = calculate_sid(sr_beginning_label, sid_index)
        }

	// Creating and upserting ls_node documents  
	parse_ls_node_router(a, bgp_id, name, router_ip, srgb, sr_node_sid)
	parse_ls_node_internal_router(a, bgp_id, name, router_ip, srgb, igp_router_id, sr_node_sid)
	parse_ls_node_internal_transport_prefix(a, bgp_id, name, router_ip, srgb, sr_prefix_sid)
}


func parse_srgb(ls_sr string) string {
	// Transforming ls_sr to SRGB
        srgb_split := strings.Split(ls_sr, " ")
        srgb_start, srgb_range := srgb_split[2], srgb_split[1]
        combining_srgb := []string{srgb_start, srgb_range}
        combined_srgb  := strings.Join(combining_srgb, ", ")
	return combined_srgb
}

func parse_sr_beginning_label(srgb string) int {
        srgb_split := strings.Split(srgb, ", ")
        sr_beginning_label := srgb_split[0]
        sr_beginning_label_val, _ := strconv.ParseInt(sr_beginning_label, 10, 0)
        return int(sr_beginning_label_val)
}

// Parses a Router from the current LSNode OpenBMP message
// Upserts the created Router document into the Routers collection
func parse_ls_node_router(a *ArangoHandler, bgp_id string, name string, router_ip string, srgb string, sr_node_sid string) {
        fmt.Println("Parsing ls_node - document 1: router_document")
        router_document := &database.Router{
                BGPID:     bgp_id,
                Name:      name,
                RouterIP:  router_ip,
		SRNodeSID: sr_node_sid,
                SRGB:      srgb,
        }
        if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("While upserting the current ls_node message's router document, encountered an error", err)
        } else {
        	fmt.Printf("Successfully added current ls_node message's router document -- Router: %q with SRGB: %q, SRNodeSID: %q, and name: %q\n", router_ip, srgb, sr_node_sid, name)
	}
}

// Parses an Internal Router from the current LSNode OpenBMP message
// Upserts the created Internal Router document into the Routers collection
func parse_ls_node_internal_router(a *ArangoHandler, bgp_id string, name string, router_ip string, srgb string, igp_router_id string, sr_node_sid string) {
        fmt.Println("Parsing ls_node - document 2: internal_router_document")
        internal_router_document := &database.InternalRouter{
                BGPID:     bgp_id,
                Name:      name,
                RouterIP:  router_ip,
		SRNodeSID: sr_node_sid,
                SRGB:      srgb,
		IGPID:     igp_router_id,
        }
        if err := a.db.Upsert(internal_router_document); err != nil {
                fmt.Println("While upserting the current ls_node message's internal router document, encountered an error", err)
        } else {
        	fmt.Printf("Successfully added current ls_node message's internal router document -- Internal Router: %q with SRGB: %q, SRNodeSID: %q, and name: %q\n", router_ip, srgb, sr_node_sid, name)
	}	
}

// Parses an Internal Transport Prefix from the current LSNode OpenBMP message
// Upserts the created Internal Transport Prefix document into the Routers collection
func parse_ls_node_internal_transport_prefix(a *ArangoHandler, bgp_id string, name string, router_ip string, srgb string, sr_prefix_sid string) {
        fmt.Println("Parsing ls_node - document 3: internal_transport_prefix_document")
        internal_transport_prefix_document := &database.InternalTransportPrefix{
                BGPID:       bgp_id,
                Name:        name,
                RouterIP:    router_ip,
		SRPrefixSID: sr_prefix_sid,
                SRGB:        srgb,
        }
        if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("While upserting the current ls_node message's internal transport prefix document, encountered an error", err)
        } else {
        	fmt.Printf("Successfully added current ls_node message's internal transport prefix document -- Internal Transport Prefix: %q with SRGB: %q, SRPrefixSID: %q, and name: %q\n", router_ip, srgb, sr_prefix_sid, name)
	}
}
