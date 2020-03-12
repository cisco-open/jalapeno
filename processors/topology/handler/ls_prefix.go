package handler

import (
	"fmt"
	"strings"
	"wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/database"
	"wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/openbmp"
)

func ls_prefix(a *ArangoHandler, m *openbmp.Message) {
        if m.Action() != "add" {
                fmt.Println("Action was not 'add' -- not parsing ls_prefix message")
                return
        }

	// Collecting necessary fields from message
        prefix         := m.GetStr("prefix")
        ls_prefix_sid  := m.GetStr("ls_prefix_sid")
	router_ip      := prefix
	bgp_id	       := prefix

	if prefix == "" {
		fmt.Println("No Prefix for current ls_prefix message -- skipping")
	}
        if ls_prefix_sid == "" {
                fmt.Println("No SIDIndex for the current ls_prefix message -- skipping")
                return
        }

	node_sid_index := parse_sid_index(ls_prefix_sid)

	// Collecting potentially existing SR initial label from previously upserted documents -- this may be empty
	sr_beginning_label := a.db.GetSRBeginningLabel(bgp_id)
	sr_node_sid := ""
	if(sr_beginning_label != 0) {
		sr_node_sid = calculate_sid(sr_beginning_label, node_sid_index)
	}

        // Creating and upserting peer documents
        parse_ls_prefix_router(a, bgp_id, router_ip, node_sid_index, sr_node_sid)
        parse_ls_prefix_internal_router(a, bgp_id, router_ip, node_sid_index, sr_node_sid)
}


// Parses sid index from ls_prefix_sid field
func parse_sid_index(ls_prefix_sid string) string {
        sid_split := strings.Split(ls_prefix_sid, " ")
        sid_index := sid_split[len(sid_split)-1]
	return sid_index
}


// Parses a Router from the current LS-Prefix OpenBMP message
// Upserts the created Router document into the Routers collection
func parse_ls_prefix_router(a *ArangoHandler, bgp_id string, router_ip string, node_sid_index string, sr_node_sid string) {
        fmt.Println("Parsing peer - document: router_document")
        router_document := &database.Router{
                BGPID:        bgp_id,
                RouterIP:     router_ip,
		NodeSIDIndex: node_sid_index,
                SRNodeSID:    sr_node_sid,
        }
	if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("While upserting the current peer message's router document, encountered an error:", err)
        } else {
                fmt.Printf("Successfully added current peer message's router document: Router: %q with NodeSIDIndex: %q and SRNodeSID: %q\n", router_ip, node_sid_index, sr_node_sid)
        }
}


// Parses a Internal Router from the current LS-Prefix OpenBMP message
// Upserts the created Internal Router document into the InternalRouters collection
func parse_ls_prefix_internal_router(a *ArangoHandler, bgp_id string, router_ip string, node_sid_index string, sr_node_sid string) {
        fmt.Println("Parsing ls_prefix - document 2: internal_router_document")
        internal_router_document := &database.InternalRouter{
                BGPID:        bgp_id,
                RouterIP:     router_ip,
		NodeSIDIndex: node_sid_index,
                SRNodeSID:    sr_node_sid,
        }
        if err := a.db.Upsert(internal_router_document); err != nil {
                fmt.Println("While upserting the current ls_prefix message's internal router document, encountered an error:", err)
        } else {
                fmt.Printf("Successfully added current ls_prefix message's internal router document -- Internal Router: %q with NodeSIDIndex: %q and SRNodeSID: %q\n", router_ip, node_sid_index, sr_node_sid)
        }
}


