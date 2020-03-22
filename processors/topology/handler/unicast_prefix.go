package handler

import (
//        "strings"
        "fmt"
        "wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/database"
        "wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/openbmp"
)

func unicast_prefix(a *ArangoHandler, m *openbmp.Message) {
        // Collecting necessary fields from message
	peer_ip              :=  m.GetStr("peer_ip")
        peer_asn                  :=  m.GetStr("peer_asn")
        prefix               :=  m.GetStr("prefix")
        prefix_len           :=  m.GetStr("prefix_len")
        is_ipv4              :=  m.GetStr("is_ipv4")
        as_path              :=  m.GetStr("as_path")
        as_path_count        :=  m.GetStr("as_path_count")
        origin_as            :=  m.GetStr("origin_as")
        nexthop              :=  m.GetStr("nexthop")
        med                  :=  m.GetStr("med")
        local_pref           :=  m.GetStr("local_pref")
        community_list       :=  m.GetStr("community_list")
        ext_community_list   :=  m.GetStr("ext_community_list")
        is_nexthop_ipv4      :=  m.GetStr("is_nexthop_ipv4")
        labels               :=  m.GetStr("labels")

	// Creating and upserting epe_node documents
	parse_epe_prefix(a, peer_ip, peer_asn, prefix, prefix_len, is_ipv4, as_path, as_path_count, origin_as, nexthop, med, local_pref, community_list, ext_community_list, is_nexthop_ipv4, labels)
}

// Parses an EPEPrefix from the current unicast_prefix OpenBMP message
// Upserts the created EPEPrefix document into the EPEPrefix vertex collection
func parse_epe_prefix(a *ArangoHandler, peer_ip string, peer_asn string, prefix string, prefix_len string, is_ipv4 string, as_path string, as_path_count string, origin_as string, nexthop string, med string, local_pref string, community_list string, ext_community_list string, is_nexthop_ipv4 string, labels string) {
        fmt.Println("Parsing epe_prefix - document 1: epe_prefix_document")

        peer_has_internal_asn :=  check_asn_location(peer_asn)

        // case 1: neighboring peer is internal -- this is not an EPE prefix
        if peer_asn == a.asn || peer_has_internal_asn == true {
                fmt.Println("Current peer message's neighbor ASN is a local ASN: this is not an EPEPrefix -- skipping")
                return
        }

	epe_prefix_document := &database.EPEPrefix{
                PeerIP:        peer_ip,
		PeerASN:       peer_asn,
		Prefix:        prefix,
		Length:        prefix_len,
		Nexthop:       nexthop,
		ASPath:        as_path,
		OriginASN:     origin_as,
		ASPathCount:   as_path_count,
		MED:           med,
		LocalPref:     local_pref,
		CommunityList: community_list,
		ExtComm:       ext_community_list,
		IsIPv4:        is_ipv4,
		IsNexthopIPv4: is_nexthop_ipv4,
		Labels:        labels,

        }
        if err := a.db.Upsert(epe_prefix_document); err != nil {
                fmt.Println("Encountered an error while upserting the current unicast_prefix message's epe_prefix document", err)
        } else {
		fmt.Printf("Successfully added epe_prefix document -- EPEPrefix: %q, peer IP: %q, with prefix: %q\n", peer_ip, prefix)
        }
}

