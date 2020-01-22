package handler

import (
	"fmt"
        "strings"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/jalapeno/services/collectors/topology/openbmp"
)

func peer(a *ArangoHandler, m *openbmp.Message) {
        if m.Action() != openbmp.ActionUp {
                fmt.Println("Action was down -- not adding peer")
                return
        }

        // Collecting necessary fields from message
        local_bgp_id     := m.GetStr("local_bgp_id")
	local_router_ip  := local_bgp_id
        local_asn        := m.GetStr("local_asn")
        local_intf_ip    := m.GetStr("local_ip")
        remote_bgp_id    := m.GetStr("remote_bgp_id")
	remote_router_ip := remote_bgp_id
        remote_asn       := m.GetStr("remote_asn")
        remote_intf_ip   := m.GetStr("remote_ip")
	
	// Creating and upserting peer documents
        parse_peer_router(a, local_bgp_id, local_router_ip, local_asn)
        parse_peer_router(a, remote_bgp_id, remote_router_ip, remote_asn)
        parse_peer_internal_router(a, local_bgp_id, local_router_ip, local_asn)
	parse_peer_internal_router(a, remote_bgp_id, remote_router_ip, remote_asn)

        parse_peer_border_router(a, local_bgp_id, local_router_ip, local_asn, remote_asn)
	parse_peer_border_router(a, remote_bgp_id, remote_router_ip, remote_asn, local_asn)

        parse_peer_external_router(a, local_bgp_id, local_router_ip, local_asn)
	parse_peer_external_router(a, remote_bgp_id, remote_router_ip, remote_asn)
        parse_peer_internal_transport_prefix(a, local_bgp_id, local_router_ip, local_asn)
       	parse_peer_internal_transport_prefix(a, remote_bgp_id, remote_router_ip, remote_asn)

        parse_peer_router_interface(a, local_bgp_id, local_router_ip, local_intf_ip, local_asn, remote_asn)
        parse_peer_router_interface(a, remote_bgp_id, remote_router_ip, remote_intf_ip, remote_asn, local_asn)

}


// Parses a Router from the current Peer OpenBMP message
// Upserts the created Router document into the Routers collection
func parse_peer_router(a *ArangoHandler, bgp_id string, router_ip string, asn string) {
	fmt.Println("Parsing peer - document: router_document")

        direct_peer_asns := strings.Split(a.direct_peer_asns, " ")
        transit_provider_asns := strings.Split(a.transit_provider_asns, " ")
        peer_type := ""
        for _, element := range transit_provider_asns {
            if element == asn {
                peer_type = "Transit"
            }
        }
        for _, element := range direct_peer_asns {
            if element == asn {
                peer_type = "Direct"
            }
        }

        router_document := &database.Router{
		BGPID:    bgp_id,
		RouterIP: router_ip,
		ASN:      asn,
                PeeringType: peer_type,
	}
	if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("While upserting the current peer message's router document, encountered an error:", err)
        } else {
                fmt.Printf("Successfully added current peer message's router document: Router: %q with ASN: %q\n", router_ip, asn)
        }
}


// Parses an Internal Router from the current Peer OpenBMP message
// Upserts the created Internal Router document into the InternalRouters collection
func parse_peer_internal_router(a *ArangoHandler, bgp_id string, router_ip string, asn string) {
        fmt.Println("Parsing ls_node - document: internal_router_document")
	is_internal_asn :=  check_asn_location(asn)
	if asn != a.asn && is_internal_asn == false {
		fmt.Println("Current peer message's ASN is not local ASN: this is not an Internal Router -- skipping")
		return
	}
        internal_router_document := &database.InternalRouter{
                BGPID:    bgp_id,
                RouterIP: router_ip,
		ASN:      asn,
        }
	if err := a.db.Upsert(internal_router_document); err != nil {
                fmt.Println("While upserting the current peer message's internal router document, encountered an error", err)
        } else {
                fmt.Printf("Successfully added current peer message's internal router document -- Internal Router: %q with ASN: %q\n", router_ip, asn)
        }
}


// Parses a Border Router from the current Peer OpenBMP message
// Upserts the created Border Router document into the BorderRouters collection
func parse_peer_border_router(a *ArangoHandler, bgp_id string, router_ip string, src_asn string, dst_asn string) {
        fmt.Println("Parsing ls_node - document: border_router_document")
	src_has_internal_asn :=  check_asn_location(src_asn)
	dst_has_internal_asn :=  check_asn_location(dst_asn)

	// case 1: neighboring peer is internal -- this is not a border router
	// case 2: neighboring peer is external, but local node is also external -- this is not a border router
	if dst_asn == a.asn || dst_has_internal_asn == true {
		fmt.Println("Current peer message's neighbor ASN is a local ASN: this is not a Border Router -- skipping")
		return
	} else if ((dst_asn != a.asn) && (dst_has_internal_asn == false)) && ((src_asn != a.asn) || (src_has_internal_asn == false)) {
		fmt.Println("Current peer message has external ASN for both local and neighbor: this is not a Border Router -- skipping")
	}

        border_router_document := &database.BorderRouter{
                BGPID:    bgp_id,
                RouterIP: router_ip,
		ASN:      src_asn,
        }
	if err := a.db.Upsert(border_router_document); err != nil {
                fmt.Println("While upserting the current peer message's border router document, encountered an error", err)
        } else {
                fmt.Printf("Successfully added current peer message's border router document -- Border Router: %q with ASN: %q\n", router_ip, src_asn)
        }
}


// Parses an External Router from the current Peer OpenBMP message
// Upserts the created External Router document into the ExternalRouters collection
func parse_peer_external_router(a *ArangoHandler, bgp_id string, router_ip string, asn string) {
        fmt.Println("Parsing ls_node - document: external_router_document")
	is_internal_asn :=  check_asn_location(asn)
	if asn == a.asn || is_internal_asn == true {
		fmt.Println("Current peer message's ASN is local ASN: this is an Internal Router -- skipping")
		return
	}

        direct_peer_asns := strings.Split(a.direct_peer_asns, " ")
        transit_provider_asns := strings.Split(a.transit_provider_asns, " ")

        peer_type := "None"        
        for _, element := range transit_provider_asns {
            if element == asn {
                peer_type = "Transit"
            } 
        }
        for _, element := range direct_peer_asns {
            if element == asn {
                peer_type = "Direct"
            }
        }

        external_router_document := &database.ExternalRouter{
                BGPID:    bgp_id,
                RouterIP: router_ip,
		ASN:      asn,
                PeeringType: peer_type,
        }
	if err := a.db.Upsert(external_router_document); err != nil {
                fmt.Println("While upserting the current peer message's external router document, encountered an error", err)
        } else {
                fmt.Printf("Successfully added current peer message's external router document -- External Router: %q with ASN: %q and Peer Type: %q\n", router_ip, asn, peer_type)
        }
}


// Parses an Internal Transport Prefix from the current Peer OpenBMP message
// Upserts the created Internal Transport Prefix document into the InternalTransportPrefixes collection
func parse_peer_internal_transport_prefix(a *ArangoHandler, bgp_id string, router_ip string, asn string) {
        fmt.Println("Parsing peer - document: internal_transport_prefix_document")
	is_internal_asn :=  check_asn_location(asn)
	if asn != a.asn && is_internal_asn == false {
		fmt.Println("Current peer message's ASN is not local ASN: this is not an Internal Transport Prefix -- skipping")
		return
	}
	internal_transport_prefix_document := &database.InternalTransportPrefix{
		BGPID:    bgp_id,
		RouterIP: router_ip,
		ASN:      asn,
	}
	if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("While upserting the current peer message's internal transport prefix document, encountered an error")
	} else {
                fmt.Printf("Successfully added current peer message's internal transport prefix document -- Internal Transport Prefix: %q with ASN: %q\n", router_ip, asn)
        }
}


// Parses a Router Interface from the current Peer OpenBMP message
// Upserts the created Router Interface document into either the BorderRouterInterfaces collection 
// or the ExternalRouterInterfaces collection depending on the Router location
func parse_peer_router_interface(a *ArangoHandler, bgp_id string, router_ip string, router_intf_ip string, router_asn string, peer_router_asn string) {
        fmt.Println("Parsing peer - document: router interface document")

	router_has_internal_asn :=  check_asn_location(router_asn)
	peer_router_has_internal_asn :=  check_asn_location(peer_router_asn)

	// We don't want to parse router-interfaces when the Peer OpenBMP message is from internal-router to internal-router
	// This is because the internal bgp relationship is described in Peer OpenBMP messages as loopback to loopback
	// We will instead parse these interfaces from LSLink OpenBMP messages
	if (router_asn == a.asn || router_has_internal_asn) && (peer_router_asn == a.asn || peer_router_has_internal_asn) {
		fmt.Println("Skipping parsing current peer message's router interface document -- internal to internal, will parse in ls_link")
		return
	}

	if router_asn != a.asn && router_has_internal_asn == false {
		external_router_interface_document := &database.ExternalRouterInterface {
			BGPID:             bgp_id,
			RouterIP:          router_ip,
			RouterInterfaceIP: router_intf_ip,
			RouterASN:         router_asn,
		}
		if err := a.db.Upsert(external_router_interface_document); err != nil {
        	        fmt.Println("While upserting the current peer message's external router interface document, encountered an error")
		} else {
        	        fmt.Printf("Successfully added current peer message's external router interface document -- External Router Interface: %q with ASN: %q and Interface: %q\n", router_ip, router_asn, router_intf_ip)
	        }
		parse_peer_external_prefix_edge(a, router_ip, router_asn, router_intf_ip)
	} else {
		border_router_interface_document := &database.BorderRouterInterface {
			BGPID:             bgp_id,
			RouterIP:          router_ip,
			RouterInterfaceIP: router_intf_ip,
			RouterASN:         router_asn,
		}
		if err := a.db.Upsert(border_router_interface_document); err != nil {
        	        fmt.Println("While upserting the current peer message's border router interface document, encountered an error")
		} else {
        	        fmt.Printf("Successfully added current peer message's border router interface document -- Border Router Interface: %q with ASN: %q and Interface: %q\n", router_ip, router_asn, router_intf_ip)
	        }
	}
}

// Parses an External Prefix Edge from the current Peer OpenBMP message
// Upserts the created External Prefix Edge document into the ExternalPrefixEdges collection
func parse_peer_external_prefix_edge(a *ArangoHandler, router_ip string, router_asn string, router_intf_ip string) {
        fmt.Println("Parsing peer - document: external_prefix_edge_document")
        a.db.CreateExternalPrefixEdgeSource(router_ip, router_asn, router_intf_ip)
}
