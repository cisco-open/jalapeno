package handler

import (
	"strings"
        "fmt"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
)

func ls_link(a *ArangoHandler, m *openbmp.Message) {
	// Collecting necessary fields from message
        src_router_id    :=  m.GetStr("router_id")
        src_interface_ip :=  m.GetStr("intf_ip")
	src_asn          :=  m.GetStr("local_node_asn")
        dst_router_id    :=  m.GetStr("remote_router_id")
        dst_interface_ip :=  m.GetStr("nei_ip")
	dst_asn 	 :=  m.GetStr("remote_node_asn")
        protocol         :=  m.GetStr("protocol")
        link_label       :=  m.GetStr("ls_adjacency_sid")
	epe_label        :=  m.GetStr("peer_node_sid")

	// End parsing if core fields are missing
        if (link_label == "") && (epe_label == "") {
                fmt.Println("No ls_adjacency_sid or peer_node_sid available, skipping all ls_link parsing for this message")
                return
        }

        // Creating and upserting ls_node documents
	// internal-router to internal-router
	if (link_label != "") {
		link_label_message := strings.Split(link_label, " ")
        	adjacency_sid := link_label_message[len(link_label_message)-1]
		parse_ls_link_internal_link_edge(a, src_router_id, src_interface_ip, dst_router_id, dst_interface_ip, protocol, adjacency_sid)		
		parse_ls_link_internal_router_interface(a, src_router_id, src_interface_ip, src_asn, adjacency_sid)
		parse_ls_link_internal_router_interface(a, dst_router_id, dst_interface_ip, dst_asn, adjacency_sid)
	} 
	// internal-router to internal-bgp-only-router or to external-router
	if (epe_label != "") {
		epe_label_message := strings.Split(epe_label, " ")
	        epe_sid := epe_label_message[len(epe_label_message)-1]
	        src_has_internal_asn := check_asn_location(src_asn)
	        dst_has_internal_asn := check_asn_location(dst_asn)
		// internal-router to internal-bgp-only-router
		if((src_asn == a.asn) || (src_has_internal_asn)) && ((dst_asn == a.asn) || (dst_has_internal_asn)) {
			parse_ls_link_internal_link_edge(a, src_router_id, src_interface_ip, dst_router_id, dst_interface_ip, protocol, epe_sid)
			parse_ls_link_internal_router_interface(a, src_router_id, src_interface_ip, src_asn, epe_sid)
			parse_ls_link_internal_router_interface(a, dst_router_id, dst_interface_ip, dst_asn, epe_sid)
		} else { 	// internal-router to external-router
			parse_ls_link_external_link_edge(a, src_router_id, src_interface_ip, dst_router_id, dst_interface_ip, protocol, epe_sid)
			if((src_asn == a.asn) || src_has_internal_asn) {
				parse_ls_link_border_router_interface(a, src_router_id, src_interface_ip, src_asn, epe_sid)
				parse_ls_link_external_router_interface(a, dst_router_id, dst_interface_ip, dst_asn)		
				parse_ls_link_external_prefix_edge(a, dst_router_id, dst_interface_ip, dst_asn)		
			} else {
				parse_ls_link_border_router_interface(a, dst_router_id, dst_interface_ip, dst_asn, epe_sid)
				parse_ls_link_external_router_interface(a, src_router_id, src_interface_ip, src_asn)
				parse_ls_link_external_prefix_edge(a, src_router_id, src_interface_ip, src_asn)		
			}
		}
	}
}


// Parses an Internal Link Edge from the current LS-Link OpenBMP message
// Upserts the created Internal Link Edge document into the InternalLinkEdges collection
func parse_ls_link_internal_link_edge(a *ArangoHandler, src_router_id string, src_interface_ip string, dst_router_id string, 
                                      dst_interface_ip string, protocol string, label string) {
        fmt.Println("Parsing ls_link - document: internal_link_edge_document")
        fmt.Printf("Parsing current ls_link message's internal_link_edge document: From Router: %q through Interface: %q and Label: %q " +
                   "to Router: %q through Interface: %q\n", src_router_id, src_interface_ip, label, dst_router_id, dst_interface_ip)
	a.db.CreateInternalLinkEdge(src_router_id, dst_router_id, src_interface_ip, dst_interface_ip, protocol, label) 

        internal_link_edge_document := &database.InternalLinkEdge{
	        SrcIP:          src_router_id,
		DstIP:	        dst_router_id,
		SrcInterfaceIP: src_interface_ip,
		DstInterfaceIP: dst_interface_ip,
		Protocol:  	protocol,
		Label:		label,
        }
	internal_link_edge_document.SetKey()
	if err := a.db.Insert(internal_link_edge_document); err != nil {
                fmt.Println("While upserting the current ls_link message's internal_link_edge document, encountered an error:", err)
        } else {
                fmt.Printf("Successfully added current ls_link message's internal_link_edge document: From Router: %q through Interface: %q and Label: %q " +
                            "to Router: %q through Interface: %q\n", src_router_id, src_interface_ip, label, dst_router_id, dst_interface_ip)
        }
}


// Parses an External Link Edge from the current LS-Link OpenBMP message
// Upserts the created External Link Edge document into the ExternalLinkEdges collection
func parse_ls_link_external_link_edge(a *ArangoHandler, src_router_id string, src_interface_ip string, dst_router_id string, 
                                      dst_interface_ip string, protocol string, epe_label string) {
        fmt.Println("Parsing ls_link - document: external_link_edge_document")
        fmt.Printf("Parsing current ls_link message's external_link_edge document: From Router: %q through Interface: %q and Label: %q " +
                   "to Router: %q through Interface: %q\n", src_router_id, src_interface_ip, epe_label, dst_router_id, dst_interface_ip)

	key := src_router_id + "_" + src_interface_ip + "_" + dst_interface_ip + "_" + dst_router_id
	external_link_edge_exists := a.db.CheckExternalLinkEdgeExists(key)
	if external_link_edge_exists {
		a.db.UpdateExternalLinkEdge(src_router_id, dst_router_id, src_interface_ip, dst_interface_ip, protocol, epe_label) 
	} else {
		a.db.CreateExternalLinkEdge(src_router_id, dst_router_id, src_interface_ip, dst_interface_ip, protocol, epe_label) 
	}

        external_link_edge_document := &database.ExternalLinkEdge{
	        SrcIP:          src_router_id,
		DstIP:	        dst_router_id,
		SrcInterfaceIP: src_interface_ip,
		DstInterfaceIP: dst_interface_ip,
		Protocol:  	protocol,
		Label:		epe_label,
        }
	external_link_edge_document.SetKey()
	if err := a.db.Insert(external_link_edge_document); err != nil {
                fmt.Println("While upserting the current ls_link message's external_link_edge document, encountered an error:", err)
        } else {
                fmt.Printf("Successfully added current ls_link message's external_link_edge document: From Router: %q through Interface: %q and Label: %q " +
                           "to Router: %q through Interface: %q\n", src_router_id, src_interface_ip, epe_label, dst_router_id, dst_interface_ip)
        }
}



// Parses an Internal Router Interface from the current LSLink OpenBMP message
// Upserts the created Internal Router Interface document into the InternalRouterInterfaces collection
func parse_ls_link_internal_router_interface(a *ArangoHandler, router_ip string, router_intf_ip string, router_asn string, router_intf_adjacency_sid string) {
        fmt.Println("Parsing ls_link - document: internal router interface document")
	bgp_id := router_ip
        internal_router_interface_document := &database.InternalRouterInterface {
                BGPID:             bgp_id,
                RouterIP:          router_ip,
                RouterInterfaceIP: router_intf_ip,
                RouterASN:         router_asn,
		AdjacencyLabel:    router_intf_adjacency_sid,
        }
        if err := a.db.Upsert(internal_router_interface_document); err != nil {
        	fmt.Println("While upserting the current ls-link message's internal router interface document, encountered an error")
        } else {
               	fmt.Printf("Successfully added current ls-link message's internal router interface document -- Internal Router Interface: %q with ASN: %q and Interface: %q and AdjacencyLabel: %q\n", router_ip, router_asn, router_intf_ip, router_intf_adjacency_sid)
        }
}

// Parses a Border Router Interface from the current LSLink OpenBMP message
// Upserts the created Border Router Interface document into the BorderRouterInterfaces collection
func parse_ls_link_border_router_interface(a *ArangoHandler, router_ip string, router_intf_ip string, router_asn string, router_intf_epe_sid string) {
        fmt.Println("Parsing ls_link - document: border router interface document")
	bgp_id := router_ip
        border_router_interface_document := &database.BorderRouterInterface {
                BGPID:             bgp_id,
                RouterIP:          router_ip,
                RouterInterfaceIP: router_intf_ip,
                RouterASN:         router_asn,
		EPELabel:          router_intf_epe_sid,
        }
        if err := a.db.Upsert(border_router_interface_document); err != nil {
        	fmt.Println("While upserting the current ls-link message's border router interface document, encountered an error")
        } else {
               	fmt.Printf("Successfully added current ls-link message's border router interface document -- Border Router Interface: %q with ASN: %q and Interface: %q and AdjacencyLabel: %q\n", router_ip, router_asn, router_intf_ip, router_intf_epe_sid)
        }
}

// Parses an External Router Interface from the current LSLink OpenBMP message
// Upserts the created External Router Interface document into the ExternalRouterInterfaces collection
func parse_ls_link_external_router_interface(a *ArangoHandler, router_ip string, router_intf_ip string, router_asn string) {
        fmt.Println("Parsing ls_link - document: external router interface document")
	bgp_id := router_ip
        external_router_interface_document := &database.ExternalRouterInterface {
                BGPID:             bgp_id,
                RouterIP:          router_ip,
                RouterInterfaceIP: router_intf_ip,
                RouterASN:         router_asn,
        }
        if err := a.db.Upsert(external_router_interface_document); err != nil {
        	fmt.Println("While upserting the current ls-link message's external router interface document, encountered an error")
        } else {
               	fmt.Printf("Successfully added current ls-link message's external router interface document -- External Router Interface: %q with ASN: %q and Interface: %q\n", router_ip, router_asn, router_intf_ip)
        }
}


// Parses an External Prefix Edge from the current LSLink OpenBMP message
// Upserts the created External Prefix Edge document into the ExternalPrefixEdges collection
func parse_ls_link_external_prefix_edge(a *ArangoHandler, router_ip string, router_intf_ip string, router_asn string) {
        fmt.Println("Parsing ls_link - document: external_prefix_edge_document")

	// In this function, we are not creating a new ExternalPrefixEdge. 
	// There's a chance that an ExternalPrefixEdge with the source ExternalRouter already exists in the 
        // ExternalPrefixEdges collection. This is due to parsing of unicast-prefix messages. However, that 
        // ExternalPrefixEdge document will have only the source ExternalRouter's ASN and InterfaceIP, not 
        // the router's IP itself. This function will get all ExternalPrefixEdges with "SrcInftIP" == router_intf_ip
        // and "SrcRouterASN" == router_asn. It will then update "SrcRouterIP" and "_from" with router_ip parsed here.

	// If no ExternalPrefixEdges exist with the currently parsed router_intf_ip and router_asn -- that's okay. Previously,
	// an ExternalPrefixEdge document would have been made with just the source ExternalRouter aspects parsed
	// in this ls-link message: specifically router_ip, router_intf_ip, and router_asn. However -- the first issue is that
	// ExternalPrefixEdge keys require the prefix destination, and keys cannot be updated, so creating an ExternalPrefixEdge
	// record without a destination prefix would lead to a broken data model. Secondly, the concern would be that the association
	// between router_ip and router_intf_ip would be lost -- however, that relationship is recorded in ExternalRouterInterfaces
	// parsed in this script itself earlier on. That relationship will be checked for when unicast-prefix messages arrive
	// with the prefix destination and other relevant data, and thus we will not lose that association. 
        a.db.CreateExternalPrefixEdgeSource(router_ip, router_asn, router_intf_ip)

}
