package handler

import (
	"fmt"
	"strings"

	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/log"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/openbmp"
)

func unicast_prefix(a *ArangoHandler, m *openbmp.Message) {
        prefix := m.GetStr("prefix")
        peer_ip := m.GetStr("peer_ip")
        peer_asn := m.GetStr("peer_asn")
        all_labels := m.GetStr("labels")
        nexthop := m.GetStr("nexthop")
        as_path := m.GetStr("as_path")

	leng, ok := m.GetInt("prefix_len")
	if !ok {
		leng = 0
	}
        log.Infof("Got to be inserted: Prefix %s/%d via %s [asn: %v]", prefix, leng, peer_ip, peer_asn)

        // Parsing a Prefix from current UnicastPrefix OpenBMP message
	prefix_document := &database.Prefix{
		Prefix: prefix,
		Length: leng,
	}
	prefix_document.SetKey()

        // Handle messages with "delete" action - TBD
	if m.Action() == openbmp.ActionDel {
		//a.db.Delete(prefix_document)
		//return
	} else if m.Action() != openbmp.ActionAdd {
		return
	}

	labels := strings.Split(all_labels, ",")
	if len(labels) == 1 && labels[0] == "" {
		labels = nil
	}

	routerKey := a.db.GetRouterKeyFromInterfaceIP(peer_ip)

	// TODO... do we add router here???
	if routerKey == "" {
		log.Warningln("Could not find router key for ", peer_ip, prefix)
		return
	}

	if peer_asn == "6500" && labels != nil {
		log.Infof("Got Prefix %s/%d from local node %s/%s... not adding (LABELS: %v)", prefix_document.Prefix, prefix_document.Length, peer_ip, peer_asn, labels)
		return
	}
	if peer_asn == a.asn || peer_asn == "6500" {
		log.Infof("Got Prefix %s/%d from local node %s/%s... not adding (LABELS: %v)", prefix_document.Prefix, prefix_document.Length, peer_ip, peer_asn, labels)
		return
	}
	a.db.Insert(prefix_document)
        prefix_documentID, err := database.GetID(prefix_document)
	if err != nil {
		fmt.Println("Could not get id?", err)
		return
	}

        // Parsing a PrefixEdge from current UnicastPrefix OpenBMP message
	prefix_edge_document := &database.PrefixEdge{
		To:   prefix_documentID,
		From: routerKey,
	}
	if a.db.Read(prefix_edge_document) != nil {
		prefix_edge_document = &database.PrefixEdge{
			NextHop:     nexthop,
			InterfaceIP: peer_ip,
			ASPath:      strings.Split(as_path, " "),
                        To:          prefix_documentID,
                        From:        routerKey,
			Labels:      labels,
		}
		if err := a.db.Insert(prefix_edge_document); err != nil {
			log.Errorln("Could not insert", err)
		}
		return
	}

	if len(labels) > 0 {
		prefix_edge_document.Labels = labels
		//log.Infof("Prefix %s --> %s Label: %s", routerKey, prefix_edge_document.InterfaceIP, prefix_edge_document.Labels)
	}
	if as_path := strings.Split(m.GetStr("as_path"), " "); len(as_path) > 0 {
		prefix_edge_document.ASPath = as_path
	}
	if err := a.db.Upsert(prefix_edge_document); err != nil {
		log.Errorln("Could not upsert", err)
		return
	}
	log.Infof("Added Prefix %s/%d via %s [asn: %v] [lbl: %s]", prefix, leng, peer_ip, peer_asn, labels)
}
