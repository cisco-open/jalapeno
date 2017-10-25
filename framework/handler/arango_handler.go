package handler

import (
	"fmt"
	"strings"
	"time"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/arango"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/openbmp"
)

type ArangoHandler struct {
	fmap map[string]HandlerFunc
	db   arango.ArangoConn
	asn  string
}

func NewArango(db arango.ArangoConn, localASN string) *ArangoHandler {
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
	t := time.Date(2017, 10, 14, 1, 0, 0, 0, time.UTC)
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
	fmt.Println(m)

	l := &arango.Router{
		BGPID:    m.GetStr("local_bgp_id"),
		ASN:      m.GetStr("local_asn"),
		RouterIP: m.GetStr("router_ip"),
		IsLocal:  false,
	}
	if l.ASN == a.asn {
		l.IsLocal = true
	}

	r := &arango.Router{
		//IP:       m.GetStr("remote_ip"),
		BGPID:   m.GetStr("remote_bgp_id"),
		ASN:     m.GetStr("remote_asn"),
		IsLocal: false,
	}
	if r.ASN == a.asn {
		r.IsLocal = true
	}

	if l.BGPID == "::" || r.BGPID == "::" {
		return
	}

	if err := a.db.Upsert(l); err != nil {
		//log.WithError(err).Error("Error on upserting router")
	}

	if err := a.db.Insert(r); err != nil {
		//log.WithError(err).Error("Error on inserting router")
	}

	rID, err := arango.GetID(r)
	if err != nil {
		log.WithError(err).Error("Could not get To ID")
		return
	}
	lID, err := arango.GetID(l)
	if err != nil {
		log.WithError(err).Error("Could not get From ID")
		return
	}

	ed := &arango.LinkEdge{
		To:     rID,
		From:   lID,
		FromIP: m.GetStr("local_ip"),
		ToIP:   m.GetStr("remote_ip"),
	}

	if ed.FromIP == l.BGPID && ed.ToIP == r.BGPID {
		log.Warningf("Not sure if I should add this link: %+v", ed)
		return
	}

	if err := a.db.Insert(ed); err != nil {
	}
	ed = &arango.LinkEdge{
		From:   rID,
		To:     lID,
		ToIP:   m.GetStr("local_ip"),
		FromIP: m.GetStr("remote_ip"),
	}
	if err := a.db.Insert(ed); err != nil {
	}

	log.Infof("Router %v/%v (%v) --> (%v) Peer %v/%v ", l.BGPID, l.ASN, ed.FromIP, ed.ToIP, r.BGPID, r.ASN)
}

func (a *ArangoHandler) HandleCollector(m *openbmp.Message) {
	fmt.Println(m)
	if m.Action() != openbmp.ActionHeartbeat {
		log.Debugf("Got Collector %s [seq %v] action: %v.\n", m.GetUnsafe("admin_id"), m.GetUnsafe("sequence"), m.Action())
	}
}

func (a *ArangoHandler) HandleBaseAttribute(m *openbmp.Message) {

}

func (a *ArangoHandler) HandleUnicastPrefix(m *openbmp.Message) {
	leng, ok := m.GetInt("prefix_len")
	if !ok {
		leng = 0
	}

	p := &arango.Prefix{
		Prefix: m.GetStr("prefix"),
		Length: leng,
	}
	p.SetKey()
	if m.Action() == openbmp.ActionDel {
		//a.db.Delete(p)
		//return
	} else if m.Action() != openbmp.ActionAdd {
		return
	}
	labels := strings.Split(m.GetStr("labels"), ",")
	if len(labels) == 1 && labels[0] == "" {
		labels = nil
	}

	rKey := a.db.GetRouterKeyFromInterfaceIP(m.GetStr("peer_ip"))
	// TODO... do we add router here???
	if rKey == "" {
		log.Warningln("Could not find router key for ", m.GetStr("peer_ip"), m.GetStr("prefix"))
		return
	}

	if m.GetStr("peer_asn") == a.asn || m.GetStr("peer_asn") == "6500" {
		log.Debugf("Got Prefix %s/%d from local node %s/%s... not adding", p.Prefix, p.Length, m.GetStr("peer_ip"), m.GetStr("peer_asn"))
		return
	}
	a.db.Insert(p)
	pID, err := arango.GetID(p)
	if err != nil {
		return
	}

	ed := &arango.PrefixEdge{
		To:   pID,
		From: rKey,
	}

	if a.db.Read(ed) != nil {
		ed = &arango.PrefixEdge{
			NextHop:     m.GetStr("nexthop"),
			InterfaceIP: m.GetStr("peer_ip"),
			ASPath:      strings.Split(m.GetStr("as_path"), " "),
			To:          pID,
			From:        rKey,
			Labels:      labels,
		}
		a.db.Insert(ed)
		return
	}
	if len(labels) > 0 {
		ed.Labels = labels
		log.Debugf("Prefix %s --> %s Label: %s", rKey, ed.InterfaceIP, ed.Labels)
	}
	if as_path := strings.Split(m.GetStr("as_path"), " "); len(as_path) > 0 {
		ed.ASPath = as_path
	}
	a.db.Upsert(ed)
	log.Infof("Added Prefix %s/%d via %s [asn: %v]", m.GetStr("prefix"), leng, m.GetStr("peer_ip"), m.GetStr("peer_asn"))
}

func (a *ArangoHandler) HandleLSNode(m *openbmp.Message) {
	fmt.Println(m)
	return
}

func (a *ArangoHandler) HandleLSLink(m *openbmp.Message) {
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
	f := &arango.Router{
		BGPID:    m.GetOneOfIP("router_id", "peer_ip"),
		ASN:      m.GetOneOf("peer_asn", "local_node_asn"),
		RouterIP: m.GetStr("router_ip"),
		IsLocal:  false,
	}
	if strings.Contains(f.BGPID, ":") {
		//TODO... why did we even get this?
		return
	}
	if f.ASN == a.asn {
		f.IsLocal = true
	}

	f.SetKey()
	if a.db.Insert(f) == nil {
		inserted = true
	}

	// TODO: Do I try to add this guy too?
	t := &arango.Router{
		BGPID: m.GetOneOfIP("remote_router_id"),
		ASN:   m.GetStr("remote_node_asn"),
	}
	t.SetKey()

	l := &arango.LinkEdge{
		ToIP:   m.GetOneOfIP("nei_ip", "peer_ip"),
		FromIP: m.GetOneOfIP("intf_ip", "router_ip"),
		Label:  lbl,
	}
	l.SetEdge(t, f)
	l.SetKey()
	if inserted || lbl != "" {
		a.db.Upsert(l)
	} else {
		a.db.Insert(l)
	}
	log.Infof("Added Link: %v_%v(%v) [%s] -->  %v_%v(%v) [%s]: Labels: %v", f.BGPID, f.ASN, f.RouterIP, l.FromIP, t.BGPID, t.ASN, t.RouterIP, l.ToIP, lbl)
}

func (a *ArangoHandler) HandleBMPStat(m *openbmp.Message) {
	log.Debugln(m.String())
}

func (a *ArangoHandler) HandleLSPrefix(m *openbmp.Message) {
	fmt.Println(m)
}

func (a *ArangoHandler) HandleRouter(m *openbmp.Message) {
	log.Debugln(m.String())
}