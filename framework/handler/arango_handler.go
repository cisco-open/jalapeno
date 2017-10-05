package handler

import (
	"strings"
	"time"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/arango"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/openbmp"
)

type ArangoHandler struct {
	fmap map[string]HandlerFunc
	db   arango.ArangoConn
}

func NewArango(db arango.ArangoConn) *ArangoHandler {
	a := &ArangoHandler{
		fmap: make(map[string]HandlerFunc),
		db:   db,
	}
	a.fmap[openbmp.TopicPeer] = a.HandlePeer
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
	t := time.Date(2017, 10, 3, 12, 0, 0, 0, time.Now().Location())
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
		log.WithField("Action", m.Action()).Warnf("Unsupported Action type")
		return
	}
	log.Debug(m)
	l := &arango.Router{
		BGPID:    m.GetStr("local_bgp_id"),
		ASN:      m.GetStr("local_asn"),
		RouterIP: m.GetStr("router_ip"),
	}

	r := &arango.Router{
		//IP:       m.GetStr("remote_ip"),
		BGPID: m.GetStr("remote_bgp_id"),
		ASN:   m.GetStr("remote_asn"),
	}

	if err := a.db.Upsert(l); err != nil {
		log.WithError(err).Error("Error on upserting router")
	}
	if err := a.db.Insert(r); err != nil {
	}

	ed := &arango.LinkEdge{
		To:     r.GetID(),
		From:   l.GetID(),
		FromIP: m.GetStr("local_ip"),
		ToIP:   m.GetStr("remote_ip"),
	}

	if err := a.db.Insert(ed); err != nil {
	}

	ed = &arango.LinkEdge{
		From:   r.GetID(),
		To:     l.GetID(),
		ToIP:   m.GetStr("local_ip"),
		FromIP: m.GetStr("remote_ip"),
	}
	if err := a.db.Insert(ed); err != nil {
	}

	log.Infof("Added: \n\tRouter %v\n\tEdge %v\n\tRouter %v\n\n", l.GetKey(), ed.GetKey(), r.GetKey())
}

func (a *ArangoHandler) HandleCollector(m *openbmp.Message) {
	if m.Action() != openbmp.ActionHeartbeat {
		log.Debugf("Got Collector %s [seq %v] action: %v.\n", m.GetUnsafe("admin_id"), m.GetUnsafe("sequence"), m.Action())
	}
}

func (a *ArangoHandler) HandleBaseAttribute(m *openbmp.Message) {
	log.Debug(m)
}

func (a *ArangoHandler) HandleUnicastPrefix(m *openbmp.Message) {
	log.Debug(m)
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
		a.db.Delete(p)
		return
	} else if m.Action() != openbmp.ActionAdd {
		log.WithField("Action", m.Action()).Warn("Not Delete or Add")
		return
	}

	a.db.Insert(p)
	rKey := a.db.GetRouterKeyFromInterfaceIP(m.GetStr("peer_ip"))
	// TODO... do we add router here???
	if rKey == "" {
		if rKey = a.db.GetRouterKeyFromInterfaceIP(m.GetStr("nexthop")); rKey == "" {
			rKey = arango.Router{
				ASN:   m.GetStr("peer_asn"),
				BGPID: m.GetStr("peer_ip"),
			}.GetID()
		}
	}
	labels := strings.Split(m.GetStr("labels"), ",")
	if len(labels) == 1 && labels[0] == "" {
		labels = nil
	}
	ed := &arango.PrefixEdge{
		To:   p.GetID(),
		From: rKey,
	}

	if a.db.Read(ed) != nil {
		ed = &arango.PrefixEdge{
			NextHop:     m.GetStr("nexthop"),
			InterfaceIP: m.GetStr("peer_ip"),
			ASPath:      strings.Split(m.GetStr("as_path"), " "),
			To:          p.GetID(),
			From:        rKey,
			Labels:      labels,
		}
		a.db.Insert(ed)
		return
	}
	if len(labels) > 0 {
		ed.Labels = labels
	}
	if as_path := strings.Split(m.GetStr("as_path"), " "); len(as_path) > 0 {
		ed.ASPath = as_path
	}
	a.db.Upsert(ed)
}

func (a *ArangoHandler) HandleLSNode(m *openbmp.Message) {
	log.Debug(m)
	return
}

func (a *ArangoHandler) HandleLSLink(m *openbmp.Message) {
	if m.Action() != openbmp.ActionAdd {
		return
	}
	log.Debug(m)
	lbl := m.GetOneOf("ls_adjacency_sid", "peer_node_sid")
	if lbl == "" {
		return
	}
	lbls := strings.Split(lbl, " ")
	lbl = lbls[len(lbls)-1]
	inserted := false
	f := &arango.Router{
		BGPID:    m.GetStr("local_router_id"),
		ASN:      m.GetStr("local_node_asn"),
		RouterIP: m.GetStr("router_ip"),
	}
	if a.db.Insert(f) == nil {
		inserted = true
	}
	t := &arango.Router{
		BGPID:    m.GetStr("remote_router_id"),
		ASN:      m.GetStr("remote_node_asn"),
		RouterIP: m.GetStr("peer_ip"),
	}

	l := &arango.LinkEdge{
		ToIP:   m.GetOneOf("nei_ip", "peer_ip"),
		FromIP: m.GetOneOf("intf_ip", "router_ip"),
		Label:  lbl,
	}
	l.SetEdge(t, f)
	if inserted || lbl != "" {
		a.db.Upsert(l)
	} else {
		a.db.Insert(l)
	}
}

func (a *ArangoHandler) HandleBMPStat(m *openbmp.Message) {
	log.Debug(m)
}

func (a *ArangoHandler) HandleLSPrefix(m *openbmp.Message) {
	log.Debug(m)
}
