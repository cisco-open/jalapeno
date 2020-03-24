package handler

import (
	"fmt"
	"github.com/cisco-ie/jalapeno/processors/topology/database"
	"github.com/cisco-ie/jalapeno/processors/topology/openbmp"
)

func peer(a *ArangoHandler, m *openbmp.Message) {
        if m.Action() != openbmp.ActionUp {
                fmt.Println("Action was down -- not adding peer")
                return
        }

        // Collecting necessary fields from message
        local_bgp_id     := m.GetStr("local_bgp_id")
        local_asn        := m.GetStr("local_asn")
        remote_asn       := m.GetStr("remote_asn")
        peer_ip          := m.GetStr("remote_ip")

	// Creating and upserting peer documents
        parse_peer_epe_node(a, local_bgp_id, peer_ip, local_asn, remote_asn)

}

// Parses an EPE Node from the current Peer OpenBMP message
// Updates entries in EPENode collection
func parse_peer_epe_node(a *ArangoHandler, local_bgp_id string, peer_ip string, local_asn string, remote_asn string) {
        fmt.Println("Parsing peer - document: epe_node_document")

	var peer_list [] string
        peer_list = append(peer_list, peer_ip)

        remote_has_internal_asn :=  check_asn_location(remote_asn)

        // case 1: neighboring peer is internal -- this is not a border router
        if remote_asn == a.asn || remote_has_internal_asn == true {
                fmt.Println("Current peer message's neighbor ASN is a local ASN: this is not an EPENode -- skipping")
                return
        }

        epe_node_exists := a.db.CheckExistingEPENode(local_bgp_id)
        if (epe_node_exists) {
            tempVariable := a.db.GetExistingPeerIP(peer_ip)
            print(tempVariable)
            a.db.UpdateExistingPeerIP(local_bgp_id, peer_ip)
        } else {

        epe_node_document := &database.EPENode{
                RouterID:  local_bgp_id,
		PeerIP:    peer_list,
	}
        if err := a.db.Upsert(epe_node_document); err != nil {
                fmt.Println("Encountered an error while upserting the epe node document", err)
        } else {
                fmt.Printf("Successfully added epe node document: %q with peer: %q\n", local_bgp_id, peer_ip)
        }
    }
}
