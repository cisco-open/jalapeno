package handler

import (
	"strings"

	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/log"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/openbmp"
)

func ls_link(a *ArangoHandler, m *openbmp.Message) {
	if strings.Contains(m.String(), "0.2.0.2") { // Bruce is working on removing this.
		return
	}
	if m.Action() != openbmp.ActionAdd {
		return
	}

	lbl := m.GetOneOf("ls_adjacency_sid", "peer_node_sid")
	lbls := strings.Split(lbl, " ")
	lbl = lbls[len(lbls)-1]
	inserted := false

        // Parsing a Router from current LSLink OpenBMP message
	router_document := &database.Router{
		BGPID:    m.GetOneOfIP("router_id", "peer_ip"),
		ASN:      m.GetOneOf("peer_asn", "local_node_asn"),
		RouterIP: m.GetStr("router_ip"),
		IsLocal:  false,
	}
	if strings.Contains(router_document.BGPID, ":") {
		//TODO... why did we even get this?
		//return
	}
	if router_document.ASN == a.asn {
		router_document.IsLocal = true
                log.Infof("Router has a local ASN!")
	}
	router_document.SetKey()
	if a.db.Insert(router_document) == nil {
		inserted = true
	}

	// TODO: Do I try to add this guy too?
	router_document2 := &database.Router{
		BGPID: m.GetOneOfIP("remote_router_id"),
		ASN:   m.GetStr("remote_node_asn"),
	}
	router_document2.SetKey()

        // Parsing a LinkEdge from current LSLink OpenBMP message
	link_edge_document := &database.LinkEdge{
		ToIP:   m.GetOneOfIP("nei_ip", "peer_ip"),
		FromIP: m.GetOneOfIP("intf_ip", "router_ip"),
		Label:  lbl,
	}
	if link_edge_document.Label == "" && (strings.Contains(link_edge_document.ToIP, ":") || strings.Contains(link_edge_document.FromIP, ":")) {
		// TODO: IF ipv6 with no label... don't add. Is this what we want??
		//return
	}
	link_edge_document.SetEdge(router_document2, router_document)
	link_edge_document.SetKey()
	if inserted || lbl != "" {
		a.db.Upsert(link_edge_document)
	} else {
		a.db.Insert(link_edge_document)
	}
	log.Infof("Added Link: %v_%v(%v) [%s] -->  %v_%v(%v) [%s]: Labels: %v", router_document.BGPID, router_document.ASN, router_document.RouterIP, link_edge_document.FromIP, router_document2.BGPID, router_document2.ASN, router_document2.RouterIP, link_edge_document.ToIP, lbl)
}
