package handler

import (
	"fmt"
	"strings"
	"time"
        "strconv"

	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/log"
	"wwwin-github.cisco.com/spa-ie/voltron/services/collectors/topology/openbmp"
)

type ArangoHandler struct {
	fmap map[string]HandlerFunc
	db   database.ArangoConn
	asn  string
}

func NewArango(db database.ArangoConn, localASN string) *ArangoHandler {
	a := &ArangoHandler{
		fmap: make(map[string]HandlerFunc),
		db:   db,
		asn:  localASN,
	}
	a.fmap[openbmp.TopicPeer] = a.HandlePeer
	a.fmap[openbmp.TopicRouter] = a.HandleRouter
	a.fmap[openbmp.TopicCollector] = a.HandleCollector
	a.fmap[openbmp.TopicBaseAttribute] = a.HandleBaseAttribute
	a.fmap[openbmp.TopicBMPStat] = a.HandleBMPStat
	a.fmap[openbmp.TopicUnicastPrefix] = a.HandleUnicastPrefix
	a.fmap[openbmp.TopicLSNode] = a.HandleLSNode
	a.fmap[openbmp.TopicLSLink] = a.HandleLSLink
	a.fmap[openbmp.TopicLSPrefix] = a.HandleLSPrefix
	return a
}

func (a *ArangoHandler) Handle(m *openbmp.Message) {
	ts, ok := m.GetTimestamp()
	t := time.Date(2017, 11, 16, 1, 0, 0, 0, time.UTC)
	if !ok || ts.Before(t) {
		return
	}

	if f, ok := a.fmap[m.Topic.String()]; ok {
		f(m)
	} else {
		log.WithField("Topic", m.Topic.String()).Warn("Unknown topic")
	}
}

func (a *ArangoHandler) Register(topic openbmp.Topic, f HandlerFunc) {
	a.fmap[topic.String()] = f
}

func (a *ArangoHandler) RegisterDefault(f HandlerFunc) {
	log.Debugf("Register Default")
}

func (a *ArangoHandler) HandlePeer(m *openbmp.Message) {
	if m.Action() != openbmp.ActionUp {
		return
	}
        log.Infof("Handling HandlePeer")
        fmt.Println("Handling Peer OpenBMP message")
        fmt.Println(m)

        local_bgp_id := m.GetStr("local_bgp_id")
        local_asn := m.GetStr("local_asn")
        router_ip := m.GetStr("router_ip")
        remote_bgp_id := m.GetStr("remote_bgp_id")
        remote_asn := m.GetStr("remote_asn")
        local_ip := m.GetStr("local_ip")
        remote_ip := m.GetStr("remote_ip")

        // Parsing a Router document from current Peer OpenBMP message
        router_document := &database.Router{
		BGPID:    local_bgp_id,
		ASN:      local_asn,
		RouterIP: router_ip,
		IsLocal:  false,
	}
	if router_document.ASN == a.asn {
		router_document.IsLocal = true
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
		ASN:     remote_asn,
		IsLocal: false,  
	}
	if router_document2.ASN == a.asn {
		router_document2.IsLocal = true
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
		To:     router_document2ID,
		From:   router_documentID,
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
		ToIP:   local_ip,
		FromIP: remote_ip,
	}
	if err := a.db.Insert(link_edge_document2); err != nil {
	}
	log.Infof("Router %v/%v (%v) --> (%v) Peer %v/%v ", router_document.BGPID, router_document.ASN, link_edge_document2.FromIP, link_edge_document2.ToIP, router_document2.BGPID, router_document2.ASN)

        // Parsing an Internal Transport Prefix from current Peer OpenBMP message
	internal_transport_prefix_document := &database.InternalTransportPrefix{
		BGPID:    local_bgp_id,
		ASN:      local_asn,
		RouterIP: router_ip,
		IsLocal:  false,
	}
	if internal_transport_prefix_document.ASN == a.asn {
		internal_transport_prefix_document.IsLocal = true
	}
	if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("While upserting the current message's internal transport prefix document, encountered an error")
	}
}

func (a *ArangoHandler) HandleCollector(m *openbmp.Message) {
        log.Infof("Handling HandleCollector")
        fmt.Println("Handling Collector OpenBMP message")
	fmt.Println(m)
	if m.Action() != openbmp.ActionHeartbeat {
		log.Debugf("Got Collector %s [seq %v] action: %v.\n", m.GetUnsafe("admin_id"), m.GetUnsafe("sequence"), m.Action())
	}
}

func (a *ArangoHandler) HandleBaseAttribute(m *openbmp.Message) {
        log.Infof("Handling HandleBaseAttribute")
        fmt.Println("Handling Base Attribute OpenBMP message")
        fmt.Println(m)
}

func (a *ArangoHandler) HandleUnicastPrefix(m *openbmp.Message) {
        log.Infof("Handling HandleUnicastPrefix")
        fmt.Println("Handling UnicastPrefix OpenBMP message")
        fmt.Println(m)

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

func (a *ArangoHandler) HandleLSNode(m *openbmp.Message) {
        log.Infof("Handling HandleLSNode")
        fmt.Println("Handling LSNode OpenBMP message")
        fmt.Println(m)

        ls_sr := m.GetStr("ls_sr_capabilities")
        // fmt.Println("Current message has ls_sr:, s)

        srgb_split := strings.Split(ls_sr, " ")
        srgb_start, srgb_range := srgb_split[2], srgb_split[1]
        // fmt.Println("Current message has srgb_start:", srgb_start)
        // fmt.Println("Current message has srgb_range:", srgb_range)

        name := m.GetStr("name")
        router_id := m.GetStr("router_id")
        combining_srgb := []string{srgb_start, srgb_range}
        combined_srgb := strings.Join(combining_srgb, ", ")
        // fmt.Println("Current message has srgb:", combined_srgb)
        
        // routerKey := a.db.GetRouterByIP(m.GetStr("router_id"))
        // fmt.Println("Current message has routerKey:", routerKey)

        // Parsing a Router from current LSNode OpenBMP message
	router_document := &database.Router{
                BGPID: router_id,
		SRGB: combined_srgb,
                Name: name,
	}
	if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("While upserting the current message's router document, encountered an error", err)
                return
	}
        fmt.Println("Successfully added Router:", router_id, "with SRGB:", combined_srgb, "and name:", name)

        // Parsing an Internal Transport Prefix from current LSNode OpenBMP message
	internal_transport_prefix_document := &database.InternalTransportPrefix{
                BGPID: router_id,
		SRGB: combined_srgb,
                Name: name,
	}
	if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("While upserting the current message's internal transport prefix document, encountered an error", err)
                return
	}
        fmt.Println("Successfully added Internal Transport Prefix:", router_id, "with SRGB:", combined_srgb, "and name:", name)
}

func (a *ArangoHandler) HandleLSLink(m *openbmp.Message) {
        log.Infof("Handling HandleLSLink")
        fmt.Println("Handling LSLink OpenBMP message")
        fmt.Println(m)

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

func (a *ArangoHandler) HandleBMPStat(m *openbmp.Message) {
        log.Infof("Handling HandleBMPStat")
        fmt.Println("Handling BMPStat OpenBMP message")
	fmt.Println(m)
}

func (a *ArangoHandler) HandleLSPrefix(m *openbmp.Message) {
        log.Infof("Handling HandleLSPrefix")
        fmt.Println("Handling LSPrefix OpenBMP message")
	fmt.Println(m)

        ls_prefix_sid := m.GetStr("ls_prefix_sid")
        prefix := m.GetStr("prefix")

        if ls_prefix_sid == "" {
                fmt.Println("No SRNodeSID for the current ls_prefix message")
                return
        }

        // an assumption is made here that the SR-node-SID index is always the last value in the ls_prefix_sid output (field 34) from ls_prefix OpenBMP messages 
        // for example, in "(ls_prefix_sid): SPF 1001" and "(ls_prefix_sid): N SPF 1", the index is the last value (1001 and 1 respectively).
        sr_node_sid_split := strings.Split(ls_prefix_sid, " ")
        sr_node_sid_index := sr_node_sid_split[len(sr_node_sid_split)-1]
        
        // this is all data manipulation
        // for example, to go from an "ls_prefix_sid" of "N SPF 1", to a "sr_node_sid" of "16001" (assuming the beginning label is 16000)
        srgb := a.db.GetSRBeginningLabel(prefix)
        // fmt.Println("We got sr_beginning_label:", srgb)
        if srgb == "" {
                fmt.Println("No SRGB for the current ls_prefix message")
                return
        }
        srgb_split := strings.Split(srgb, ", ")
        // fmt.Println("We got SRGB for the current ls_prefix message:", srgb_split)
        sr_beginning_label := srgb_split[0]
        // fmt.Println("We got SRGB Beginning Label:", sr_beginning_label)
        sr_beginning_label_val, _ := strconv.ParseInt(sr_beginning_label, 10, 0)
        sr_node_sid_index_val, _ := strconv.ParseInt(sr_node_sid_index, 10, 0)
        sr_node_sid_val := sr_beginning_label_val + sr_node_sid_index_val
        sr_node_sid := strconv.Itoa(int(sr_node_sid_val))
        // fmt.Println("Parsed SRNodeSID:", sr_node_sid)

        // Parsing a Router from current LSPrefix OpenBMP message
        router_document := &database.Router{
                BGPID: prefix,
                SRNodeSID: sr_node_sid,
        }
        if err := a.db.Upsert(router_document); err != nil {
                fmt.Println("Something went wrong with ls_prefix for a router:", err)
                return
        }
        fmt.Println("Successfully added Router:", prefix, "with SRNodeSID:", sr_node_sid)

        // Parsing an Internal Transport Prefix from current LSPrefix OpenBMP message
        internal_transport_prefix_document := &database.InternalTransportPrefix{
                BGPID: prefix,
                SRNodeSID: sr_node_sid,
        }
        if err := a.db.Upsert(internal_transport_prefix_document); err != nil {
                fmt.Println("Something went wrong with ls_prefix for an internal transport prefix:", err)
                return
        }
        fmt.Println("Successfully added Internal Transport Prefix:", prefix, "with SRNodeSID:", sr_node_sid)
}

func (a *ArangoHandler) HandleRouter(m *openbmp.Message) {
        log.Infof("Handling HandleRouter")
        fmt.Println("Handling Router OpenBMP message")
	fmt.Println(m)
}
