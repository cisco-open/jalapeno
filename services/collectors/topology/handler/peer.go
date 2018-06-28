package handler

import (
	"fmt"

	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/log"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/openbmp"
)

func peer(a *ArangoHandler, m *openbmp.Message) {
        if m.Action() != openbmp.ActionUp {
                fmt.Println("Action was down -- not adding peer")
                return
        }

        local_bgp_id := m.GetStr("local_bgp_id")
        local_ip := m.GetStr("local_ip")
        local_asn := m.GetStr("local_asn")

        remote_bgp_id := m.GetStr("remote_bgp_id")
        remote_ip := m.GetStr("remote_ip")
        remote_asn := m.GetStr("remote_asn")


        // Parsing a Router document from current Peer OpenBMP message
        router_document := &database.Router{
		BGPID:    local_bgp_id,
		IsLocal:  false,
		ASN:      local_asn,
	}
	if router_document.ASN == a.asn {
		router_document.IsLocal = true
                fmt.Println("Router has local ASN!")
	}
	if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("While upserting the current message's router document, encountered an error")
	}
	router_documentID, err := database.GetID(router_document)
	if err != nil {
		log.WithError(err).Error("Could not get From ID")
		return
	}


        // Parsing a second Router document from current Peer OpenBMP message
	router_document2 := &database.Router{
		BGPID:   remote_bgp_id,
		IsLocal: false,  
		ASN:     remote_asn,
	}
	if router_document2.ASN == a.asn {
		router_document2.IsLocal = true
                fmt.Println("Router2 has local ASN!")
	}
	if err := a.db.Insert(router_document2); err != nil {
                fmt.Println("While upserting the current message's router2 document, encountered an error")
	}
	router_document2ID, err := database.GetID(router_document2)
	if err != nil {
		log.WithError(err).Error("Could not get To ID")
		return
	}


        // Parsing a Link-Edge document from current Peer OpenBMP message
	link_edge_document := &database.LinkEdge{
		From:   router_documentID,
		To:     router_document2ID,
		FromIP: local_ip,
		ToIP:   remote_ip,
	}
	// Loopbacks questionable -- should these be added?
	if link_edge_document.FromIP == router_document.BGPID && link_edge_document.ToIP == router_document2.BGPID {
		log.Warningf("Not sure if I should add this link: %+v", link_edge_document)
		return
	}
	if err := a.db.Insert(link_edge_document); err != nil {
	}


        // Parsing a second Link-Edge document from current Peer OpenBMP message
	link_edge_document2 := &database.LinkEdge{
		From:   router_document2ID,
		To:     router_documentID,
		FromIP: remote_ip,
		ToIP:   local_ip,
	}
	if err := a.db.Insert(link_edge_document2); err != nil {
	}
	log.Infof("Router %v/%v (%v) --> (%v) Peer %v/%v ", router_document.BGPID, router_document.ASN, link_edge_document2.FromIP, link_edge_document2.ToIP, router_document2.BGPID, router_document2.ASN)


        // Parsing an Internal Transport Prefix from current Peer OpenBMP message
	internal_transport_prefix_document := &database.InternalTransportPrefix{
		BGPID:    local_bgp_id,
		IsLocal:  false,
		ASN:      local_asn,
	}
	if internal_transport_prefix_document.ASN == a.asn {
		internal_transport_prefix_document.IsLocal = true
                fmt.Println("Internal Transport Prefix has local ASN!")
	}
	if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("While upserting the current message's internal transport prefix document, encountered an error")
	}
}

