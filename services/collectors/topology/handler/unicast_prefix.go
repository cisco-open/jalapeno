package handler

import (
	"fmt"
	"strings"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
)

func unicast_prefix(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        router_id         := m.GetStr("router_ip")
	prefix_ip         := m.GetStr("prefix")
        prefix            := m.GetStr("prefix")
	as_path           := m.GetStr("as_path")
	//as_path_count     := m.GetStr("as_path_count")
        asns              := strings.Split(as_path, " ")
        prefix_asn        := asns[len(asns)-1]
	prefix_length, ok := m.GetInt("prefix_len")
        sr_label          := m.GetStr("labels")
        peer_ip           := m.GetStr("peer_ip")
	peer_asn          := m.GetStr("peer_asn")
	//nexthop           := m.GetStr("nexthop")

	if !ok {
		prefix_length = 0
	}

        // Creating and upserting unicast_prefix documents
        parse_unicast_prefix_prefix(a, prefix_ip, prefix_length, prefix_asn)
	if prefix_asn == "" {
		fmt.Println("No ASN associated with unicast_prefix message -- must be internal prefix, allowing parsing")
	        parse_unicast_prefix_internal_prefix(a, prefix_ip, prefix_length, prefix_asn, sr_label)
	} else {
		is_internal_asn := check_asn_location(prefix_asn)
		if prefix_asn == a.asn || is_internal_asn {
			parse_unicast_prefix_internal_prefix(a, prefix_ip, prefix_length, prefix_asn, sr_label)
		} else {
			parse_unicast_prefix_external_prefix(a, prefix, prefix_length, prefix_asn)

			//parse_unicast_prefix_epe_topology(a, router_id, peer_ip, nexthop, prefix_asn, as_path, prefix)

			peer_has_internal_asn := check_asn_location(prefix_asn)
			//peer_has_internal_asn := check_asn_location(peer_ip)
			if (peer_asn != a.asn) && (peer_has_internal_asn == false) {
				parse_unicast_prefix_external_prefix_edge(a, peer_ip, peer_asn, prefix_ip, prefix_length, prefix_asn)
				parse_unicast_prefix_epe_prefix(a, prefix, prefix_length, router_id, peer_ip, peer_asn, prefix_asn, as_path)
			}
		}
	}
}

// Parses an EPE Prefix from the current Prefix OpenBMP message
// Upserts the created EPE Prefix document into the Prefixes collection
func parse_unicast_prefix_epe_prefix(a *ArangoHandler, prefix string, prefix_length int, router_id string, peer_ip string, peer_asn string, prefix_asn string, as_path string) {
        fmt.Println("Parsing unicast_prefix - document: epe_prefix_document")
        epe_prefix_document :=&database.EPEPrefix{
                Prefix:       prefix,
                Length:       prefix_length,
                RouterID:     router_id,
                PeerIP:       peer_ip,
		PeerASN:      peer_asn,
                RemoteASN:    prefix_asn,
                ASPath:       as_path,
	}
        epe_prefix_document.SetKey()
        //if strings.Contains(epe_prefix_document.PeerASN, "100000") {
        //        fmt.Println("This is an internal BGP advertisement of an external prefix -- skipping unicast_prefix parsing for this OpenBMP message")
        //        return
        //}
        if err := a.db.Upsert(epe_prefix_document); err != nil {
                fmt.Println("Encountered an error while upserting the current unicast_prefix epe prefix document:", err)
        } else {
                fmt.Printf("Successfully inserted current unicast_prefix message's epe prefix document: %s/%d with ASN: %v\n", prefix, prefix_length, prefix_asn)
        }
}

// Parses EPE Topology documents from unicast prefix OpenBMP message
// Upserts the document into the EPETopology edge collection
//func parse_unicast_prefix_epe_topology(a *ArangoHandler, router_id string, peer_ip string, nexthop string, prefix_asn string, as_path string, prefix string) {
//        fmt.Println("Parsing unicast_prefix - document: epe_topology_document")
//        fmt.Printf("Parsing current unicast_prefix message's epe_topology document: From EPE Node: %q through Interface: %q and Label: %q " +
//                   "to External Prefix: %q through Interface: %q\n", router_id, prefix)

//	epe_node_key := "EPENode/" + router_id
//        ext_prefix_key := "ExternalPrefix/" + prefix

//        epe_topology_document := &database.EPETopology{
//                EPENodeKey:    epe_node_key,
//                ExtPrefixKey:  ext_prefix_key,
//                RouterID:      router_id,
//                PeerIP:        peer_ip,
//		NextHop:       nexthop,
//		PrefixASN:     prefix_asn,
//                ASPath:        as_path,
//        }
//        epe_topology_document.SetKey()
//        if err := a.db.Insert(epe_topology_document); err != nil {
//                fmt.Println("Encountered an error while upserting the epe_topology document:", err)
//        } else {
//                fmt.Printf("Successfully added epe_topology document: From EPE Node: %q through Interface: %q and Label: %q " +
//                           "to External Node: %q through Interface: %q\n", router_id, prefix)
//        }
//}

// Parses a Prefix from the current Prefix OpenBMP message
// Upserts the created Prefix document into the Prefixes collection
func parse_unicast_prefix_prefix(a *ArangoHandler, prefix_ip string, prefix_length int, prefix_asn string) {
        fmt.Println("Parsing unicast_prefix - document: prefix_document")
	prefix_document := &database.Prefix{
		Prefix: prefix_ip,
		Length: prefix_length,
		ASN:	prefix_asn,
	}
	prefix_document.SetKey()
	if strings.Contains(prefix_document.Prefix, ":") {
        	fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix prefix-document parsing for this OpenBMP message")
		return
        } 
	if err := a.db.Upsert(prefix_document); err != nil {
                fmt.Println("While upserting the current unicast_prefix message's prefix document, encountered an error", err)
        } else {
                fmt.Printf("Successfully inserted current unicast_prefix message's prefix document -- Prefix %s/%d with ASN: %v\n", prefix_ip, prefix_length, prefix_asn)
	}
}

// Parses an Internal Prefix from the current Prefix OpenBMP message
// Upserts the created Internal Prefix document into the Prefixes collection
func parse_unicast_prefix_internal_prefix(a *ArangoHandler, prefix_ip string, prefix_length int, prefix_asn string, sr_label string) {
        fmt.Println("Parsing unicast_prefix - document: internal_prefix_document")
        internal_prefix_document :=&database.InternalPrefix{
                Prefix:  prefix_ip,
                Length:  prefix_length,
                ASN:     prefix_asn,
                SRLabel: sr_label,
        }
        internal_prefix_document.SetKey()
        if strings.Contains(internal_prefix_document.Prefix, ":") {
        	fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix internal-prefix-document parsing for this OpenBMP message")
		return
        } 
        if err := a.db.Upsert(internal_prefix_document); err != nil {
        	fmt.Println("While upserting the current unicast_prefix message's internal prefix document, encountered an error:", err)
        } else {
                fmt.Printf("Successfully inserted current unicast_prefix message's internal prefix document -- InternalPrefix %s/%d with ASN: %v and SR Label %s\n", prefix_ip, prefix_length, prefix_asn, sr_label)
        }
}

// Parses an External Prefix from the current Prefix OpenBMP message
// Upserts the created External Prefix document into the Prefixes collection
func parse_unicast_prefix_external_prefix(a *ArangoHandler, prefix string, prefix_length int, prefix_asn string) {
        fmt.Println("Parsing unicast_prefix - document: external_prefix_document")
        external_prefix_document :=&database.ExternalPrefix{
                Prefix:  prefix,
                Length:  prefix_length,
		ASN:     prefix_asn,
        }
        external_prefix_document.SetKey()
        if strings.Contains(external_prefix_document.Prefix, ":") {
		fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix external-prefix-document parsing for this OpenBMP message")
		return
	}
        if err := a.db.Upsert(external_prefix_document); err != nil {
		fmt.Println("While upserting the current unicast_prefix message's external prefix document, encountered an error:", err)
        } else {
		fmt.Printf("Successfully inserted current unicast_prefix message's external prefix document: %s/%d with ASN: %v\n", prefix, prefix_length, prefix_asn)
        }
}


// Parses an External Prefix Edge from the current Prefix OpenBMP message
// Upserts the created External Prefix Edge document into the ExternalPrefixEdges collection
func parse_unicast_prefix_external_prefix_edge(a *ArangoHandler, peer_ip string, peer_asn string, prefix_ip string, prefix_length int, prefix_asn string) {
        fmt.Println("Parsing unicast_prefix - document: external_prefix_edge_document")
        if strings.Contains(prefix_ip, ":") {
		fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix external-prefix-edge document parsing for this OpenBMP message")
		return
	}

	// Many unicast_prefix messages come with a "peer_ip" field that has the interface IP of the peer, instead of   
        // the IP address of the peer itself. An attempt can be made to get the actual IP of the peer by checking if 
        // the peer exists in the ExternalRouterInterfaces collection. There is no need to check if the peer exists in
        // InternalRouterInterfaces, as ExternalPrefixEdge parsing is only for edges from  ExternalRouters to ExternalPrefixes.

	// An assumption is made here as well - that two routers in the same ASN will not have identical interface-ips
	// For example, in ASN 7100, Router71-A (10.0.0.71) and Router71-B (10.0.0.72) cannot both have interface-ip 2.2.2.71.
	real_peer_ip := a.db.GetExternalRouterIP(peer_ip)
        a.db.CreateExternalPrefixEdge(real_peer_ip, peer_asn, peer_ip, prefix_ip, prefix_asn, prefix_length)

}



