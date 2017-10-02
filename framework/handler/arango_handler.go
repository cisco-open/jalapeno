package handler

import (
	"fmt"
	"strings"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/arango"
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
	if f, ok := a.fmap[m.Topic.String()]; ok {
		f(m)
	} else {
		fmt.Println("uhhhh...???", m.Topic.String())
	}
}

func (a *ArangoHandler) Register(topic openbmp.Topic, f HandlerFunc) {
	a.fmap[topic.String()] = f
}

func (a *ArangoHandler) RegisterDefault(f HandlerFunc) {

}

func (a *ArangoHandler) HandlePeer(m *openbmp.Message) {
	if m.Action().String() == openbmp.ActionDel || m.Action().String() == openbmp.ActionDown { // TODO: handle down.
		return
	} else if m.Action() != openbmp.ActionUp {
		fmt.Println("Not Down or Add... ", m.Action())
		return
	}

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
		fmt.Println("Error upserting l: ", err)
		return
	}
	if err := a.db.Insert(r); err != nil {
		//fmt.Println("Error inserting r: ", err)
		//return
	}

	ed := &arango.LinkEdge{
		To:     fmt.Sprintf("%s/%s", r.GetType(), r.GetKey()),
		From:   fmt.Sprintf("%s/%s", l.GetType(), l.GetKey()),
		FromIP: m.GetStr("local_ip"),
		ToIP:   m.GetStr("remote_ip"),
	}
	if err := a.db.Upsert(ed); err != nil {
		fmt.Printf("Erroring Inserting edge %v: %v\n", ed.GetKey(), err)
		return
	}
	fmt.Printf("Added: \n\tRouter %v\n\tEdge %v\n\tRouter %v\n\n", l.GetKey(), ed.GetKey(), r.GetKey())
}

func (a *ArangoHandler) HandleCollector(m *openbmp.Message) {
	if m.Action() != openbmp.ActionHeartbeat {
		fmt.Printf("Got Collector %s [seq %v] action: %v.\n", m.GetUnsafe("admin_id"), m.GetUnsafe("sequence"), m.Action())
	}
}

func (a *ArangoHandler) HandleBaseAttribute(m *openbmp.Message) {
	fmt.Println("BaseAttr")
	//fmt.Println(m)
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
		a.db.Delete(p)
		return
	} else if m.Action() != openbmp.ActionAdd {
		fmt.Println("Not Delete or Add... ", m.Action())
		return
	}

	a.db.Insert(p)
	r := a.db.GetRouterByIP(m.GetStr("router_ip"))
	// TODO... do we add router here???
	if r == nil {
		fmt.Println("Unable to find router...", m)
		return
	}
	r.SetKey()
	ed := &arango.PrefixEdge{
		NextHop: m.GetStr("nexthop"),
		ASPath:  strings.Split(m.GetStr("as_path"), " "),
		To:      fmt.Sprintf("%s/%s", p.GetType(), p.GetKey()),
		From:    fmt.Sprintf("%s/%s", r.GetType(), r.GetKey()),
	}

	ed.SetKey()
	if err := a.db.Upsert(ed); err != nil {
		fmt.Println("Error adding prefix ", m.GetStr("prefix"), err, ed)
		return
	}
}

func (a *ArangoHandler) HandleLSNode(m *openbmp.Message) {
	fmt.Println("lsNode", m)
	return

}

func (a *ArangoHandler) HandleLSLink(m *openbmp.Message) {
	fmt.Println("lsLink", m)
}

func (a *ArangoHandler) HandleBMPStat(m *openbmp.Message) {
	fmt.Println("bmpStat")
	//fmt.Println(m)
}

func (a *ArangoHandler) HandleLSPrefix(m *openbmp.Message) {
	fmt.Println("lsPrefix", m)
	//fmt.Println(m)
}

func (a *ArangoHandler) Debug() {
}
