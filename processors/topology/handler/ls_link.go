package handler

import (
	"strings"
        "fmt"
        "github.com/cisco-ie/jalapeno/processors/topology/database"
        "github.com/cisco-ie/jalapeno/processors/topology/openbmp"
)

func ls_link(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        local_router_id      :=  m.GetStr("router_id")
        local_interface_ip   :=  m.GetStr("intf_ip")
        asn                  :=  m.GetStr("local_node_asn")
        remote_router_id     :=  m.GetStr("remote_router_id")
        remote_interface_ip  :=  m.GetStr("nei_ip")
	peer_asn             :=  m.GetStr("remote_node_asn")
        protocol             :=  m.GetStr("protocol")
	local_pref           :=  m.GetStr("local_pref")
        med                  :=  m.GetStr("med")
        nexthop              :=  m.GetStr("nexthop")
	igp_id               :=  m.GetStr("igp_router_id")
        igp_metric           :=  m.GetStr("igp_metric")
        te_metric            :=  m.GetStr("te_default_metric")
        admin_group          :=  m.GetStr("admin_group")
        max_link_bw          :=  m.GetStr("max_link_bw")
        max_resv_bw          :=  m.GetStr("max_resv_bw")
        unresv_bw            :=  m.GetStr("unresv_bw")
        link_protection      :=  m.GetStr("link_protection")
        srlg                 :=  m.GetStr("srlg")
        link_name            :=  m.GetStr("link_name")
        link_label           :=  m.GetStr("ls_adjacency_sid")
        //adjacency_sid        :=  m.GetStr("ls_adjacency_sid")
        adj_sid_tlv          :=  m.GetStr("ls_adjacency_sid")
        epe_label            :=  m.GetStr("peer_node_sid")

	// End parsing if core fields are missing
        if (link_label == "") && (epe_label == "") {
                fmt.Println("No ls_adjacency_sid or peer_node_sid available, skipping all ls_link parsing for this message")
                return
        }

        // Creating and upserting ls_link edge documents
        if (adj_sid_tlv != "") {
                parse_ls_link(a, local_router_id, local_interface_ip, asn, remote_router_id, remote_interface_ip, protocol,
                igp_id, igp_metric, te_metric, admin_group, max_link_bw, max_resv_bw, unresv_bw, link_protection, srlg, link_name,
                adj_sid_tlv)
	}

        // Creating and upserting epe_link vertex documents
	parse_epe_link(a, local_router_id, local_interface_ip, asn, remote_router_id, remote_interface_ip, peer_asn, protocol, local_pref, med, nexthop, epe_label)
}

// Parses an LSLink Edge entry from the current LS-Link OpenBMP message
// Upserts the LSLink Edge document into the LSTopology collection
func parse_ls_link(a *ArangoHandler, local_router_id string, local_interface_ip string, asn string, remote_router_id string, remote_interface_ip string, protocol string,
igp_id string, igp_metric string, te_metric string, admin_group string, max_link_bw string, max_resv_bw string, unresv_bw string, link_protection string, srlg string, link_name string,
adj_sid_tlv string) {
        fmt.Println("Parsing ls_link message to ls_link_document")
        fmt.Printf("Parsing current ls_link message to ls_link document: From LSNode: %q through Interface: %q " +
                   "to LSNode: %q through Interface: %q\n", local_router_id, local_interface_ip, remote_router_id, remote_interface_ip)

        local_router_key := "LSNode/" + local_router_id
        remote_router_key := "LSNode/" + remote_router_id

        ls_link_document := &database.LSLink{
                LocalRouterKey:    local_router_key,
                RemoteRouterKey:   remote_router_key,
                LocalRouterID:     local_router_id,
                ASN:               asn,
                RemoteRouterID:    remote_router_id,
                LocalInterfaceIP:  local_interface_ip,
                RemoteInterfaceIP: remote_interface_ip,
                Protocol:          protocol,
                IGPID:             igp_id,
                IGPMetric:         igp_metric,
                TEMetric:          te_metric,
                AdminGroup:        admin_group,
                MaxLinkBW:         max_link_bw,
                MaxResvBW:         max_resv_bw,
                UnResvBW:          unresv_bw,
                LinkProtection:    link_protection,
                SRLG:              srlg,
                LinkName:          link_name,
                AdjacencySID:      adj_sid_tlv,
        }
        if err := a.db.Upsert(ls_link_document); err != nil {
                fmt.Println("Encountered an error while upserting ls_link document:", err)
        } else {
                fmt.Printf("Successfully added ls_link edge document from ls_link message: From Router: %q through Interface: %q " +
                            "to Router: %q through Interface: %q\n", local_router_id, local_interface_ip, remote_router_id, remote_interface_ip)
        }

        aSids := strings.Split(adj_sid_tlv, ", ")
        key := local_router_id + "_" + local_interface_ip + "_" + remote_interface_ip + "_" + remote_router_id
        for _, aSid :=  range aSids {
                s := strings.Split(aSid, " ")
                adj_sid := s[2]
                flags := s[0]
                weight := s[1]
                a.db.CreateAdjacencyList(key, adj_sid, flags, weight)
        }
}

// Parses an EPELink vertex entry from the current LS-Link OpenBMP message
// Upserts the LSLink Edge document into the LSTopology collection
func parse_epe_link(a *ArangoHandler, local_router_id string, local_interface_ip string, asn string, remote_router_id string, remote_interface_ip string, peer_asn string, protocol string, local_pref string, med string, nexthop string, epe_label string) {
        fmt.Println("Parsing ls_link message to epe_link_document")

        peer_has_internal_asn :=  check_asn_location(peer_asn)

        // case 1: neighboring peer is internal -- this is not an EPE prefix
        if peer_asn == a.asn || peer_has_internal_asn == true {
                fmt.Println("Current peer message's neighbor ASN is a local ASN: this is not an EPEPrefix -- skipping")
                return
        }

        fmt.Printf("Parsing current ls_link message to epe_link document: From EPENode: %q, Peer: %q, Interface: %q\n", local_router_id, remote_interface_ip, local_interface_ip)

        //local_router_key := "LSNode/" + local_router_id
        //remote_router_key := "LSNode/" + remote_router_id

        epe_link_document := &database.EPELink{
                //LocalRouterKey:    local_router_key,
                //RemoteRouterKey:   remote_router_key,
                RouterID:          local_router_id,
                ASN:               asn,
                PeerRouterID:      remote_router_id,
                LocalInterfaceIP:  local_interface_ip,
                PeerIP:            remote_interface_ip,
                Protocol:          protocol,
		LocalPref:         local_pref,
		MED:               med,
		Nexthop:           nexthop,
		EPELabel:          epe_label,
	}
        if err := a.db.Upsert(epe_link_document); err != nil {
                fmt.Println("Encountered an error while upserting epe_link document:", err)
        } else {
                fmt.Printf("Successfully added epe_link document from ls_link message: EPENode: %q, Peer: %q, Interface: %q\n", local_router_id, remote_interface_ip, local_interface_ip)
        }
}






