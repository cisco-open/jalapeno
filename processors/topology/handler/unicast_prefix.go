package handler

import (
	"fmt"
	"strings"
	"wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/database"
	"wwwin-github.cisco.com/spa-ie/jalapeno/processors/topology/openbmp"
)

func unicast_prefix(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        //router_id         := m.GetStr("router_ip")
	prefix_ip         := m.GetStr("prefix")
        prefix            := m.GetStr("prefix")
	//as_path           := m.GetStr("as_path")
	//as_path_count     := m.GetStr("as_path_count")
        //asns              := strings.Split(as_path, " ")
        //prefix_asn        := asns[len(asns)-1]
	prefix_length, ok := m.GetInt("prefix_len")
	if !ok {
		prefix_length = 0
	}
        label             := m.GetStr("labels")
        peer_ip           := m.GetStr("peer_ip")
	peer_asn          := m.GetStr("peer_asn")
        origin_as         := m.GetStr("origin_as")
	//nexthop           := m.GetStr("nexthop")


        // Creating and upserting unicast_prefix documents
        parse_unicast_prefix_prefix(a, prefix_ip, prefix_length, origin_as)
	if origin_as == "" {
		fmt.Println("No ASN associated with unicast_prefix message -- must be internal prefix, allowing parsing")
	        parse_unicast_prefix_internal_prefix(a, prefix_ip, prefix_length, origin_as, label)
	} else {
		is_internal_asn := check_asn_location(origin_as)
		if origin_as == a.asn || is_internal_asn {
			parse_unicast_prefix_internal_prefix(a, prefix_ip, prefix_length, origin_as, label)
		} else {
			parse_unicast_prefix_external_prefix(a, prefix, prefix_length, origin_as)

			peer_has_internal_asn := check_asn_location(origin_as)
			//peer_has_internal_asn := check_asn_location(peer_ip)
			if (peer_asn != a.asn) && (peer_has_internal_asn == false) {
				parse_unicast_prefix_external_prefix_edge(a, peer_ip, peer_asn, prefix_ip, prefix_length, origin_as)
			}
		}
	}
}

// Parses a Prefix from the current Prefix OpenBMP message
// Upserts the created Prefix document into the Prefixes collection
func parse_unicast_prefix_prefix(a *ArangoHandler, prefix_ip string, prefix_length int, origin_as string) {
        fmt.Println("Parsing unicast_prefix - document: prefix_document")
	prefix_document := &database.Prefix{
		Prefix: prefix_ip,
		Length: prefix_length,
		ASN:	origin_as,
	}
	prefix_document.SetKey()
	if strings.Contains(prefix_document.Prefix, ":") {
		fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix prefix-document parsing for this OpenBMP message")
		return
        }
	if err := a.db.Upsert(prefix_document); err != nil {
                fmt.Println("While upserting the current unicast_prefix message's prefix document, encountered an error", err)
        } else {
                fmt.Printf("Successfully inserted current unicast_prefix message's prefix document -- Prefix %s/%d with ASN: %v\n", prefix_ip, prefix_length, origin_as)
	}
}

// Parses an Internal Prefix from the current Prefix OpenBMP message
// Upserts the created Internal Prefix document into the Prefixes collection
func parse_unicast_prefix_internal_prefix(a *ArangoHandler, prefix_ip string, prefix_length int, origin_as string, label string) {
        fmt.Println("Parsing unicast_prefix - document: internal_prefix_document")
        internal_prefix_document :=&database.InternalPrefix{
                Prefix:  prefix_ip,
                Length:  prefix_length,
                ASN:     origin_as,
                Label:   label,
        }
        internal_prefix_document.SetKey()
        //if strings.Contains(internal_prefix_document.Prefix, ":") {
	//	fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix internal-prefix-document parsing for this OpenBMP message")
	//	return
        //}
        if err := a.db.Upsert(internal_prefix_document); err != nil {
		fmt.Println("Encountered an error while upserting internal prefix document:", err)
        } else {
                fmt.Printf("Successfully inserted internal prefix document -- InternalPrefix %s/%d with ASN: %v and Label %s\n", prefix_ip, prefix_length, origin_as, label)
        }
}

// Parses an External Prefix from the current Prefix OpenBMP message
// Upserts the created External Prefix document into the Prefixes collection
func parse_unicast_prefix_external_prefix(a *ArangoHandler, prefix string, prefix_length int, origin_as string) {
        fmt.Println("Parsing unicast_prefix - document: external_prefix_document")
        external_prefix_document :=&database.ExternalPrefix{
                Prefix:  prefix,
                Length:  prefix_length,
		ASN:     origin_as,
        }
        external_prefix_document.SetKey()
        //if strings.Contains(external_prefix_document.Prefix, ":") {
	//	fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix external-prefix-document parsing for this OpenBMP message")
	//	return
	//}
        if err := a.db.Upsert(external_prefix_document); err != nil {
		fmt.Println("Encountered an error while upserting external prefix document:", err)
        } else {
		fmt.Printf("Successfully inserted external prefix document: %s/%d with ASN: %v\n", prefix, prefix_length, origin_as)
        }
}


// Parses an External Prefix Edge from the current Prefix OpenBMP message
// Upserts the created External Prefix Edge document into the ExternalPrefixEdges collection
func parse_unicast_prefix_external_prefix_edge(a *ArangoHandler, peer_ip string, peer_asn string, prefix_ip string, prefix_length int, origin_as string) {
        fmt.Println("Parsing unicast_prefix - document: external_prefix_edge_document")
        //if strings.Contains(prefix_ip, ":") {
	//	fmt.Println("This is an IPv6 prefix -- skipping unicast_prefix external-prefix-edge document parsing for this OpenBMP message")
	//	return
	//}

	// Many unicast_prefix messages come with a "peer_ip" field that has the interface IP of the peer, instead of   
        // the IP address of the peer itself. An attempt can be made to get the actual IP of the peer by checking if 
        // the peer exists in the ExternalRouterInterfaces collection. There is no need to check if the peer exists in
        // InternalRouterInterfaces, as ExternalPrefixEdge parsing is only for edges from  ExternalRouters to ExternalPrefixes.

	// An assumption is made here as well - that two routers in the same ASN will not have identical interface-ips
	// For example, in ASN 7100, Router71-A (10.0.0.71) and Router71-B (10.0.0.72) cannot both have interface-ip 2.2.2.71.
	real_peer_ip := a.db.GetExternalRouterIP(peer_ip)
        a.db.CreateExternalPrefixEdge(real_peer_ip, peer_asn, peer_ip, prefix_ip, origin_as, prefix_length)

}



