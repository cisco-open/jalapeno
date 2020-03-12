package handler

import (
        "fmt"
        "strings"
	"strconv"
        "wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/database"
        "wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/openbmp"
)

func ls_node(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        router_ip     := m.GetStr("router_id")
        router_id     := m.GetStr("router_id")
	name          := m.GetStr("name")
        asn           := m.GetStr("peer_asn")
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

	parse_ls_node(a, name, router_id, asn, srgb, sr_prefix_sid, igp_router_id)
        //parse_epe_node(a, name, router_id, asn, srgb, sr_prefix_sid, igp_router_id)

	parse_ls_node_internal_router(a, bgp_id, name, router_ip, srgb, igp_router_id, sr_node_sid)
	parse_ls_node_internal_transport_prefix(a, bgp_id, name, router_ip, srgb, sr_prefix_sid)
}


func parse_srgb(ls_sr string) string {
	// Transforming ls_sr to SRGB
        srgb_split := strings.Split(ls_sr, " ")
        srgb_start, srgb_range := srgb_split[len(srgb_split)-1], srgb_split[len(srgb_split)-2]
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

// Parses an LS_Node from the current LSNode OpenBMP message
// Upserts the created LS_Node document into the LS_Node vertex collection
func parse_ls_node(a *ArangoHandler, name string, router_id string, asn string, srgb string, sr_prefix_sid string, igp_router_id string) {
        fmt.Println("Parsing ls_node - document 1: ls_node_document")
        ls_node_document := &database.LSNode{
                //BGPID:     bgp_id,
                Name:      name,
                RouterID:  router_id,
		ASN:       asn,
                PrefixSID: sr_prefix_sid,
                SRGB:      srgb,
                IGPID:     igp_router_id,
        }
        if err := a.db.Upsert(ls_node_document); err != nil {
                fmt.Println("Encountered an error while upserting the current ls_node message's ls_node document", err)
        } else {
                fmt.Printf("Successfully added ls_node document -- LSNode: %q with SRGB: %q, PrefixSID: %q, and name: %q\n", router_id, srgb, sr_prefix_sid, name)
        }
}

// Parses EPE_Node prefix-SID from an LSNode OpenBMP message
// Updates EPENode vertex documents with prefix-sid data
//func parse_epe_node(a *ArangoHandler, bgp_id string, name string, router_id string, srgb string, sr_prefix_sid string, igp_router_id string) {
//func parse_epe_node(a *ArangoHandler, name string, router_id string, asn string, srgb string, sr_prefix_sid string, igp_router_id string) {
//        fmt.Println("Parsing epe_node - document 1: epe_node_document")
//        epe_node_document := &database.EPENode{
//                //BGPID:     bgp_id,
//                Name:      name,
//                RouterID:  router_id,
//                ASN:       asn,
//                PrefixSID: sr_prefix_sid,
//                SRGB:      srgb,
//                IGPID:     igp_router_id,
//        }
//        if err := a.db.Update(epe_node_document); err != nil {
//                fmt.Println("Encountered an error while updating the current ls_node message's epe_node data", err)
//        } else {
//                fmt.Printf("Successfully added current ls_node message's epe_node data -- Router: %q with SRGB: %q, PrefixSID: %q\n", router_id, srgb, sr_prefix_sid)
//        }
//}


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
