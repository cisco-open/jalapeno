package handler

import (
	"fmt"
	"strings"
        "strconv"

	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/openbmp"
)

func ls_prefix(a *ArangoHandler, m *openbmp.Message) {
        ls_prefix_sid := m.GetStr("ls_prefix_sid")
        prefix := m.GetStr("prefix")

        if ls_prefix_sid == "" {
                fmt.Println("No SRNodeSID for the current ls_prefix message")
                return
        }

        // an assumption is made here that the SR-node-SID index is always the last value in the ls_prefix_sid output (field 34) from ls_prefix OpenBMP messages 
        // for example, in "(ls_prefix_sid): SPF 1001" and "(ls_prefix_sid): N SPF 1", the index is the last value (1001 and 1 respectively).
        sr_node_sid_split := strings.Split(ls_prefix_sid, " ")
        sr_node_sid_index := sr_node_sid_split[len(sr_node_sid_split)-1]
        
        // this is all data manipulation
        // for example, to go from an "ls_prefix_sid" of "N SPF 1", to a "sr_node_sid" of "16001" (assuming the beginning label is 16000)
        srgb := a.db.GetSRBeginningLabel(prefix)
        // fmt.Println("We got sr_beginning_label:", srgb)
        if srgb == "" {
                fmt.Println("No SRGB for the current ls_prefix message")
                return
        }
        srgb_split := strings.Split(srgb, ", ")
        // fmt.Println("We got SRGB for the current ls_prefix message:", srgb_split)
        sr_beginning_label := srgb_split[0]
        // fmt.Println("We got SRGB Beginning Label:", sr_beginning_label)
        sr_beginning_label_val, _ := strconv.ParseInt(sr_beginning_label, 10, 0)
        sr_node_sid_index_val, _ := strconv.ParseInt(sr_node_sid_index, 10, 0)
        sr_node_sid_val := sr_beginning_label_val + sr_node_sid_index_val
        sr_node_sid := strconv.Itoa(int(sr_node_sid_val))
        // fmt.Println("Parsed SRNodeSID:", sr_node_sid)

        // Parsing a Router from current LSPrefix OpenBMP message
        router_document := &database.Router{
                BGPID: prefix,
                SRNodeSID: sr_node_sid,
        }
        if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("Something went wrong with ls_prefix for a router:", err)
                return
        }
        fmt.Println("Successfully added Router:", prefix, "with SRNodeSID:", sr_node_sid)

        sr_prefix_sid := sr_node_sid
        // Parsing an Internal Transport Prefix from current LSPrefix OpenBMP message
        internal_transport_prefix_document := &database.InternalTransportPrefix{
                BGPID: prefix,
                SRPrefixSID: sr_prefix_sid,
        }
        if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("Something went wrong with ls_prefix for an internal transport prefix:", err)
                return
        }
        fmt.Println("Successfully added Internal Transport Prefix:", prefix, "with SRNodeSID:", sr_node_sid)
}

