package handler

import (
        "fmt"
        "strings"
	"strconv"
        "github.com/cisco-ie/jalapeno/processors/topology/database"
        "github.com/cisco-ie/jalapeno/processors/topology/openbmp"
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

	sr_prefix_sid := ""
        sr_beginning_label := parse_sr_beginning_label(srgb)
        sid_index := a.db.GetSIDIndex(bgp_id)
        if(sid_index != "") {
                sr_prefix_sid = calculate_sid(sr_beginning_label, sid_index)
        }

	// Creating and upserting ls_node documents  
	parse_ls_node(a, name, router_id, asn, srgb, sr_prefix_sid, igp_router_id)
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
